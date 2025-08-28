package borker

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync" // For goroutine safe access to maps
)

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
	Name   string // Field to store the client's name
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
			delete(h.channels, channel) // Clean up empty channel
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
			err := writeRESPPubSubMessage(s.writer, channel, message)
			if err != nil {
				log.Printf("Error sending message to subscriber %s (%s) on channel %s: %v", s.conn.RemoteAddr(), s.Name, channel, err)
				h.Unsubscribe(channel, s)
			}
		}
	}
}

// --- RESP Parser/Writer (unchanged) ---

// readRESP remains the same
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
		if numElements == -1 { // Null array
			return nil, nil
		}

		elements := make([]string, numElements)
		for i := 0; i < numElements; i++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			line = strings.TrimSuffix(line, "\r\n")
			if line[0] != bulkPrefix {
				return nil, fmt.Errorf("expected bulk string, got %c", line[0])
			}
			bulkLen, err := strconv.Atoi(line[1:])
			if err != nil {
				return nil, fmt.Errorf("invalid bulk string length: %w", err)
			}
			if bulkLen == -1 { // Null bulk string
				elements[i] = ""
				continue
			}

			data := make([]byte, bulkLen)
			_, err = io.ReadFull(reader, data)
			if err != nil {
				return nil, err
			}
			elements[i] = string(data)
			reader.ReadString('\n') // Read trailing \r\n
		}
		return elements, nil

	case bulkPrefix:
		bulkLen, err := strconv.Atoi(content)
		if err != nil {
			return nil, fmt.Errorf("invalid bulk string length: %w", err)
		}
		if bulkLen == -1 { // Null bulk string
			return []string{""}, nil
		}

		data := make([]byte, bulkLen)
		_, err = io.ReadFull(reader, data)
		if err != nil {
			return nil, err
		}
		reader.ReadString('\n') // Read trailing \r\n
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

// writeRESP remains the same
func writeRESP(writer *bufio.Writer, prefix byte, data string) error {
	_, err := writer.WriteString(fmt.Sprintf("%c%s\r\n", prefix, data))
	if err != nil {
		return err
	}
	return writer.Flush()
}

// writeRESPBulkString remains the same
func writeRESPBulkString(writer *bufio.Writer, data string) error {
	_, err := writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(data), data))
	if err != nil {
		return err
	}
	return writer.Flush()
}

// writeRESPArray remains the same
func writeRESPArray(writer *bufio.Writer, elements []string) error {
	_, err := writer.WriteString(fmt.Sprintf("*%d\r\n", len(elements)))
	if err != nil {
		return err
	}
	for _, elem := range elements {
		_, err = writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(elem), elem))
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

// writeRESPPubSubMessage remains the same
func writeRESPPubSubMessage(writer *bufio.Writer, channel, message string) error {
	_, err := writer.WriteString(fmt.Sprintf("*3\r\n")) // Array of 3 elements
	if err != nil {
		return err
	}

	// Element 1: "message" literal
	_, err = writer.WriteString(fmt.Sprintf("$7\r\nmessage\r\n"))
	if err != nil {
		return err
	}

	// Element 2: Channel name
	_, err = writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(channel), channel))
	if err != nil {
		return err
	}

	// Element 3: Message payload
	_, err = writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(message), message))
	if err != nil {
		return err
	}

	return writer.Flush()
}

// writeRESPSubscribeConfirmation remains the same
func writeRESPSubscribeConfirmation(writer *bufio.Writer, commandType, channel string, count int) error {
	_, err := writer.WriteString(fmt.Sprintf("*3\r\n")) // Array of 3 elements
	if err != nil {
		return err
	}

	// Element 1: commandType ("subscribe" or "unsubscribe")
	_, err = writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(commandType), commandType))
	if err != nil {
		return err
	}

	// Element 2: Channel name
	_, err = writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(channel), channel))
	if err != nil {
		return err
	}

	// Element 3: Count (integer)
	_, err = writer.WriteString(fmt.Sprintf(":%d\r\n", count))
	if err != nil {
		return err
	}

	return writer.Flush()
}

// --- Connection Handler ---

func handleConnection(conn net.Conn, hub *pubSubHub) {

	defer conn.Close()
	log.Printf("Client connected: %s", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	currentSubscriber := &subscriber{conn: conn, writer: writer}
	currentSubscriber.Name = fmt.Sprintf("client-%s", conn.RemoteAddr().String()) // Default name

	subscribedChannels := make(map[string]bool)
	isPubSubMode := false

	for {
		cmd, err := readRESP(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client %s (%s) disconnected.", conn.RemoteAddr().String(), currentSubscriber.Name)
			} else {
				log.Printf("Error reading from client %s (%s): %v", conn.RemoteAddr().String(), currentSubscriber.Name, err)
				writeRESP(writer, errorPrefix, "ERR malformed command")
			}
			break
		}

		if len(cmd) == 0 {
			writeRESP(writer, errorPrefix, "ERR empty command")
			continue
		}

		commandName := strings.ToUpper(cmd[0])
		args := cmd[1:]

		if isPubSubMode {
			switch commandName {
			case "SUBSCRIBE":
				if len(args) == 0 {
					writeRESP(writer, errorPrefix, "ERR SUBSCRIBE command requires at least one channel")
					continue
				}
				for _, channel := range args {
					hub.Subscribe(channel, currentSubscriber)
					subscribedChannels[channel] = true
					writeRESPSubscribeConfirmation(writer, "subscribe", channel, len(subscribedChannels))
				}
			case "UNSUBSCRIBE":
				if len(args) == 0 {
					for channel := range subscribedChannels {
						hub.Unsubscribe(channel, currentSubscriber)
						writeRESPSubscribeConfirmation(writer, "unsubscribe", channel, 0)
					}
					subscribedChannels = make(map[string]bool)
					isPubSubMode = false
				} else {
					for _, channel := range args {
						if subscribedChannels[channel] {
							hub.Unsubscribe(channel, currentSubscriber)
							delete(subscribedChannels, channel)
							writeRESPSubscribeConfirmation(writer, "unsubscribe", channel, len(subscribedChannels))
						}
					}
					if len(subscribedChannels) == 0 {
						isPubSubMode = false
					}
				}
			default:
				writeRESP(writer, errorPrefix, "ERR only SUBSCRIBE/UNSUBSCRIBE commands are allowed in this mode")
			}
		} else { // Normal command mode
			switch commandName {
			case "PING":
				writeRESP(writer, stringPrefix, "PONG")
			case "ECHO":
				if len(args) == 0 {
					writeRESP(writer, errorPrefix, "ERR ECHO command requires a message")
				} else {
					writeRESPBulkString(writer, args[0])
				}
			case "PUBLISH":
				if len(args) != 2 {
					writeRESP(writer, errorPrefix, "ERR PUBLISH command requires channel and message")
				} else {
					hub.Publish(args[0], args[1])
					writeRESP(writer, intPrefix, "1")
				}
			case "CLIENT":
				if len(args) == 0 {
					writeRESP(writer, errorPrefix, "ERR CLIENT command requires a subcommand")
					continue
				}
				subcommand := strings.ToUpper(args[0])
				switch subcommand {
				case "SETNAME":
					if len(args) == 2 {
						clientName := args[1]
						currentSubscriber.Name = clientName
						log.Printf("Client %s set name to: %s", conn.RemoteAddr().String(), clientName)
						writeRESP(writer, stringPrefix, "OK")
					} else {
						writeRESP(writer, errorPrefix, "ERR CLIENT SETNAME requires a name argument")
					}
				case "GETNAME": // NEW: Implement CLIENT GETNAME
					if len(args) == 1 {
						// Redis returns a bulk string with the name, or a null bulk string if no name is set
						if currentSubscriber.Name != "" {
							writeRESPBulkString(writer, currentSubscriber.Name)
						} else {
							// Null bulk string: "$-1\r\n"
							_, err := writer.WriteString("$-1\r\n")
							if err != nil {
								log.Printf("Error writing null bulk string: %v", err)
								break
							}
							writer.Flush()
						}
					} else {
						writeRESP(writer, errorPrefix, "ERR CLIENT GETNAME does not take arguments")
					}
				default:
					writeRESP(writer, errorPrefix, fmt.Sprintf("ERR Unknown CLIENT subcommand '%s'", subcommand))
				}
			case "SUBSCRIBE":
				if len(args) == 0 {
					writeRESP(writer, errorPrefix, "ERR SUBSCRIBE command requires at least one channel")
				} else {
					for _, channel := range args {
						hub.Subscribe(channel, currentSubscriber)
						subscribedChannels[channel] = true
						writeRESPSubscribeConfirmation(writer, "subscribe", channel, len(subscribedChannels))
					}
					isPubSubMode = true
				}
			default:
				writeRESP(writer, errorPrefix, fmt.Sprintf("ERR unknown command '%s'", commandName))
			}
		}
	}

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
	defer listener.Close()

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
