package borker

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

// https://chat.deepseek.com/a/chat/s/2e8611e0-2053-4feb-84b1-46474bfa0955

// --- RESP Constants ---
const (
	stringPrefix = '+'
	errorPrefix  = '-'
	intPrefix    = ':'
	bulkPrefix   = '$'
	arrayPrefix  = '*'
)

// --- Pub/Sub Hub ---
type subscriber struct {
	conn   net.Conn
	writer *bufio.Writer
	Name   string
}

type pubSubHub struct {
	mu       sync.RWMutex
	channels map[string]map[*subscriber]bool
}

func newPubSubHub() *pubSubHub {
	return &pubSubHub{
		channels: make(map[string]map[*subscriber]bool),
	}
}

func (h *pubSubHub) Subscribe(channel string, s *subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.channels[channel]; !ok {
		h.channels[channel] = make(map[*subscriber]bool)
	}
	h.channels[channel][s] = true
	log.Printf("Client %s (%s) subscribed to channel %s", s.conn.RemoteAddr(), s.Name, channel)
}

func (h *pubSubHub) Unsubscribe(channel string, s *subscriber) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.channels[channel]; ok {
		delete(subs, s)
		if len(subs) == 0 {
			delete(h.channels, channel)
		}
		log.Printf("Client %s (%s) unsubscribed from channel %s", s.conn.RemoteAddr(), s.Name, channel)
	}
}

func (h *pubSubHub) Publish(channel, message string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subs, ok := h.channels[channel]; ok {
		log.Printf("Publishing to channel '%s': '%s' (%d subscribers)", channel, message, len(subs))
		for s := range subs {
			if err := writeRESPPubSubMessage(s.writer, channel, message); err != nil {
				log.Printf("Error sending message to subscriber %s (%s) on channel %s: %v",
					s.conn.RemoteAddr(), s.Name, channel, err)
				// Note: We don't unsubscribe here as it would cause a deadlock
				// The connection cleanup will happen when the client disconnects
			}
		}
	}
}

// --- RESP Parser/Writer (optimized) ---

func readRESP(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(line, "\r\n")

	if len(line) == 0 {
		return nil, fmt.Errorf("empty line received")
	}

	prefix := line[0]
	content := line[1:]

	switch prefix {
	case arrayPrefix:
		numElements, err := strconv.Atoi(content)
		if err != nil {
			return nil, fmt.Errorf("invalid array length: %w", err)
		}
		if numElements == -1 {
			return nil, nil
		}

		elements := make([]string, numElements)
		for i := 0; i < numElements; i++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			line = strings.TrimSuffix(line, "\r\n")
			if len(line) == 0 || line[0] != bulkPrefix {
				return nil, fmt.Errorf("expected bulk string, got %s", line)
			}

			bulkLen, err := strconv.Atoi(line[1:])
			if err != nil {
				return nil, fmt.Errorf("invalid bulk string length: %w", err)
			}
			if bulkLen == -1 {
				elements[i] = ""
				continue
			}

			data := make([]byte, bulkLen)
			_, err = io.ReadFull(reader, data)
			if err != nil {
				return nil, err
			}
			elements[i] = string(data)

			// Read trailing \r\n
			_, err = reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
		}
		return elements, nil

	case bulkPrefix:
		bulkLen, err := strconv.Atoi(content)
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length: %w", err)
		}
		if bulkLen == -1 {
			return []string{""}, nil
		}

		data := make([]byte, bulkLen)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			return nil, err
		}

		// Read trailing \r\n
		_, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		return []string{string(data)}, nil

	case stringPrefix:
		return []string{content}, nil
	case errorPrefix:
		return []string{content}, fmt.Errorf("server error: %s", content)
	case intPrefix:
		return []string{content}, nil
	default:
		return nil, fmt.Errorf("unknown RESP prefix: %c", prefix)
	}
}

func writeRESP(writer *bufio.Writer, prefix byte, data string) error {
	if _, err := writer.WriteString(fmt.Sprintf("%c%s\r\n", prefix, data)); err != nil {
		return err
	}
	return writer.Flush()
}

func writeRESPBulkString(writer *bufio.Writer, data string) error {
	if _, err := writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(data), data)); err != nil {
		return err
	}
	return writer.Flush()
}

func writeRESPArray(writer *bufio.Writer, elements []string) error {
	if _, err := writer.WriteString(fmt.Sprintf("*%d\r\n", len(elements))); err != nil {
		return err
	}
	for _, elem := range elements {
		if _, err := writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem)); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func writeRESPPubSubMessage(writer *bufio.Writer, channel, message string) error {
	// Build the entire message first to minimize write operations
	fullMessage := fmt.Sprintf("*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(channel), channel, len(message), message)

	if _, err := writer.WriteString(fullMessage); err != nil {
		return err
	}
	return writer.Flush()
}

func writeRESPSubscribeConfirmation(writer *bufio.Writer, commandType, channel string, count int) error {
	// Build the entire message first to minimize write operations
	fullMessage := fmt.Sprintf("*3\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n:%d\r\n",
		len(commandType), commandType, len(channel), channel, count)

	if _, err := writer.WriteString(fullMessage); err != nil {
		return err
	}
	return writer.Flush()
}

// --- Connection Handler ---

func handleConnection(conn net.Conn, hub *pubSubHub) {
	// Ensure connection is always closed
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection from %s: %v", conn.RemoteAddr().String(), err)
		}
	}()

	log.Printf("Client connected: %s", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	currentSubscriber := &subscriber{
		conn:   conn,
		writer: writer,
		Name:   fmt.Sprintf("client-%s", conn.RemoteAddr().String()),
	}

	subscribedChannels := make(map[string]bool)
	isPubSubMode := false

	for {
		cmd, err := readRESP(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s (%s) disconnected.", conn.RemoteAddr().String(), currentSubscriber.Name)
			} else {
				log.Printf("Error reading from client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
				// Try to send error response, but ignore if it fails
				_ = writeRESP(writer, errorPrefix, "ERR malformed command")
			}
			break
		}

		if len(cmd) == 0 {
			if err := writeRESP(writer, errorPrefix, "ERR empty command"); err != nil {
				log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
				break
			}
			continue
		}

		commandName := strings.ToUpper(cmd[0])
		args := cmd[1:]

		if isPubSubMode {
			switch commandName {
			case "SUBSCRIBE":
				if len(args) == 0 {
					if err := writeRESP(writer, errorPrefix, "ERR SUBSCRIBE command requires at least one channel"); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
					continue
				}
				for _, channel := range args {
					hub.Subscribe(channel, currentSubscriber)
					subscribedChannels[channel] = true
					if err := writeRESPSubscribeConfirmation(writer, "subscribe", channel, len(subscribedChannels)); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
				}
			case "UNSUBSCRIBE":
				if len(args) == 0 {
					// Unsubscribe from all channels
					for channel := range subscribedChannels {
						hub.Unsubscribe(channel, currentSubscriber)
						if err := writeRESPSubscribeConfirmation(writer, "unsubscribe", channel, 0); err != nil {
							log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
							break
						}
					}
					subscribedChannels = make(map[string]bool)
					isPubSubMode = false
				} else {
					for _, channel := range args {
						if subscribedChannels[channel] {
							hub.Unsubscribe(channel, currentSubscriber)
							delete(subscribedChannels, channel)
							if err := writeRESPSubscribeConfirmation(writer, "unsubscribe", channel, len(subscribedChannels)); err != nil {
								log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
								break
							}
						}
					}
					if len(subscribedChannels) == 0 {
						isPubSubMode = false
					}
				}
			default:
				if err := writeRESP(writer, errorPrefix, "ERR only SUBSCRIBE/UNSUBSCRIBE commands are allowed in this mode"); err != nil {
					log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
					break
				}
			}
		} else {
			switch commandName {
			case "PING":
				if err := writeRESP(writer, stringPrefix, "PONG"); err != nil {
					log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
					break
				}
			case "ECHO":
				if len(args) == 0 {
					if err := writeRESP(writer, errorPrefix, "ERR ECHO command requires a message"); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
				} else {
					if err := writeRESPBulkString(writer, args[0]); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
				}
			case "PUBLISH":
				if len(args) != 2 {
					if err := writeRESP(writer, errorPrefix, "ERR PUBLISH command requires channel and message"); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
				} else {
					hub.Publish(args[0], args[1])
					if err := writeRESP(writer, intPrefix, "1"); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
				}
			case "CLIENT":
				if len(args) == 0 {
					if err := writeRESP(writer, errorPrefix, "ERR CLIENT command requires a subcommand"); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
					continue
				}
				subcommand := strings.ToUpper(args[0])
				switch subcommand {
				case "SETNAME":
					if len(args) == 2 {
						clientName := args[1]
						currentSubscriber.Name = clientName
						log.Printf("Client %s set name to: %s", conn.RemoteAddr().String(), clientName)
						if err := writeRESP(writer, stringPrefix, "OK"); err != nil {
							log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
							break
						}
					} else {
						if err := writeRESP(writer, errorPrefix, "ERR CLIENT SETNAME requires a name argument"); err != nil {
							log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
							break
						}
					}
				case "GETNAME":
					if len(args) == 1 {
						if currentSubscriber.Name != "" {
							if err := writeRESPBulkString(writer, currentSubscriber.Name); err != nil {
								log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
								break
							}
						} else {
							if _, err := writer.WriteString("$-1\r\n"); err != nil {
								log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
								break
							}
							if err := writer.Flush(); err != nil {
								log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
								break
							}
						}
					} else {
						if err := writeRESP(writer, errorPrefix, "ERR CLIENT GETNAME does not take arguments"); err != nil {
							log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
							break
						}
					}
				default:
					if err := writeRESP(writer, errorPrefix, fmt.Sprintf("ERR Unknown CLIENT subcommand '%s'", subcommand)); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
				}
			case "SUBSCRIBE":
				if len(args) == 0 {
					if err := writeRESP(writer, errorPrefix, "ERR SUBSCRIBE command requires at least one channel"); err != nil {
						log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
						break
					}
				} else {
					for _, channel := range args {
						hub.Subscribe(channel, currentSubscriber)
						subscribedChannels[channel] = true
						if err := writeRESPSubscribeConfirmation(writer, "subscribe", channel, len(subscribedChannels)); err != nil {
							log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
							break
						}
					}
					isPubSubMode = true
				}
			default:
				if err := writeRESP(writer, errorPrefix, fmt.Sprintf("ERR unknown command '%s'", commandName)); err != nil {
					log.Printf("Error writing to client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
					break
				}
			}
		}
	}

	// Clean up subscriptions when client disconnects
	if len(subscribedChannels) > 0 {
		log.Printf("Client %s (%s) disconnected, cleaning up subscriptions...", conn.RemoteAddr().String(), currentSubscriber.Name)
		for channel := range subscribedChannels {
			hub.Unsubscribe(channel, currentSubscriber)
		}
	}
}

func StartRespBroker() {
	hub := newPubSubHub()

	listener, err := net.Listen("tcp", ":6389")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}()

	fmt.Println("Broker with Pub/Sub and CLIENT SETNAME/GETNAME listening on :6389")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn, hub)
	}
}
