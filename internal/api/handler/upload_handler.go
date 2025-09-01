package handler

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cshum/vipsgen/vips"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mahdi-cpp/api-go-settings/internal/thumbnail"
)

//https://chat.deepseek.com/a/chat/s/913cf162-1ad1-4857-8048-2990d3c959a4

type UploadHandler struct {
	UploadDir string
}

type UploadResponse struct {
	Message  string   `json:"message"`
	Filename string   `json:"filename"`
	Size     int64    `json:"size"`
	URL      string   `json:"url"`
	Errors   []string `json:"errors,omitempty"`
}

func (h *UploadHandler) UploadJPEG(c *gin.Context) {

	vips.Startup(nil)
	defer vips.Shutdown()

	// Single file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Message: "No file uploaded",
			Errors:  []string{err.Error()},
		})
		return
	}

	// Check if it's a JPEG
	if !isJPEG(file) {
		c.JSON(http.StatusBadRequest, UploadResponse{Message: "Only JPEG files are allowed", Errors: []string{"Invalid file type"}})
		return
	}

	// Generate unique filename
	uniqueName, err := generateUniqueFilename()
	dst := filepath.Join(h.UploadDir, uniqueName)

	// Save the file
	if err := c.SaveUploadedFile(file, dst+".jpg"); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{Message: "Failed to save file", Errors: []string{err.Error()}})
		return
	}

	fmt.Println("file.Filename: ", file.Filename)
	fmt.Println("dst", dst)
	if err := thumbnail.CreateSingleThumbnail(dst); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{Message: "Failed to create thumbnail file", Errors: []string{err.Error()}})
		log.Fatalf("An error occurred during thumbnail creation: %v", err)
	}

	c.JSON(http.StatusOK, UploadResponse{
		Message:  "File uploaded successfully",
		Filename: uniqueName,
		Size:     file.Size,
		URL:      "/uploads/" + uniqueName,
	})
}

func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Message: "Failed to parse form",
			Errors:  []string{err.Error()},
		})
		return
	}

	files := form.File["files"]
	var responses []UploadResponse
	var errors []string

	for _, file := range files {
		// Check if it's a JPEG
		if !isJPEG(file) {
			errors = append(errors, fmt.Sprintf("%s: Not a JPEG file", file.Filename))
			continue
		}

		// Generate unique filename
		uniqueName, err := generateUniqueFilename()
		if err != nil {
			return
		}
		dst := filepath.Join(h.UploadDir, uniqueName)

		// Save the file
		if err := c.SaveUploadedFile(file, dst); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		responses = append(responses, UploadResponse{
			Message:  "File uploaded successfully",
			Filename: uniqueName,
			Size:     file.Size,
			URL:      "/uploads/" + uniqueName,
		})
	}

	if len(errors) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"message": "Some files failed to upload",
			"uploads": responses,
			"errors":  errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All files uploaded successfully",
		"uploads": responses,
	})
}

func (h *UploadHandler) ListFiles(c *gin.Context) {
	files, err := getJPEGFiles(h.UploadDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list files",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// Helper functions
func isJPEG(file *multipart.FileHeader) bool {
	// Check content type
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" {
		return false
	}

	// Also check file extension for extra safety
	ext := strings.ToLower(filepath.Ext(file.Filename))
	return ext == ".jpg" || ext == ".jpeg"
}

func generateUniqueFilename() (string, error) {

	u7, err2 := uuid.NewV7()
	if err2 != nil {
		return "", fmt.Errorf("error generating UUIDv7: %w", err2)
	}

	return u7.String(), nil
}

func getJPEGFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".jpg" || ext == ".jpeg" {
				// Return relative path
				rel, err := filepath.Rel(dir, path)
				if err == nil {
					files = append(files, rel)
				}
			}
		}

		return nil
	})

	return files, err
}
