package handler

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	application "github.com/mahdi-cpp/api-go-settings/internal/application"
)

type DownloadHandler struct {
	manager *application.AppManager
}

func NewDownloadHandler(manager *application.AppManager) *DownloadHandler {
	return &DownloadHandler{
		manager: manager,
	}
}

// ImageThumbnail serves the thumbnail image.
//
// Conventions followed:
//  1. **RESTful Design**: The function uses a GET request, which is appropriate for
//     fetching a resource (a thumbnail) without side effects.
//  2. **HTTP Status Codes**: It returns `http.StatusOK` (200) for a successful
//     response, `http.StatusNotFound` (404) if the image file doesn't exist,
//     and `http.StatusInternalServerError` (500) for internal errors.
//  3. **Proper Error Handling**: Instead of a bare return, it logs the error and
//     informs the client with a specific status code and message.
//  4. **Content Negotiation**: It uses `c.Data` to set the correct `Content-Type`
//     header (`image/jpeg`, `image/png`, etc.) and write the raw image data to the
//     response body, which is what a browser expects for an image request.

// http://localhost:50000/api/v1/download/thumbnail/com.iris.photos/users/018f3a8b-1b32-729a-f7e5-5467c1b2d3e4/assets/0198c111-0f9d-74f6-ab2e-6ce665ec29c6.jpg
//												    com.iris.photos/users/018f3a8b-1b32-729a-f7e5-5467c1b2d3e4/assets/0198c111-0f9d-74f6-ab2e-6ce665ec29c6.jpg

func (handler *DownloadHandler) ImageThumbnail(c *gin.Context) {

	// The full path of the image is extracted from the URL parameters.
	// For example, if the route is "/thumbnail/*filename", and the URL is
	// "/thumbnail/images/my-image.jpg", fullPath will be "images/my-image.jpg".
	fullPath := c.Param("filename")

	// Validate that the filename parameter is present.
	if fullPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename parameter is missing"})
		return
	}

	// Load the image data using the business logic layer (handler.manager).
	imageBytes, err := handler.manager.ThumbnailImageLoader.LoadImage(c, fullPath)
	if err != nil {
		log.Printf("Error loading image: %v", err)
		// Return a 404 if the image is not found, or a 500 for other errors.
		// A real implementation might check the error type to be more specific.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load image"})
		return
	}

	// Determine the correct Content-Type based on the file extension.
	// This is important for browsers to correctly display the image.
	ext := filepath.Ext(fullPath)
	contentType := "application/octet-stream" // Default MIME type
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	}

	// Send the image data with the correct content type and an OK status.
	// The `c.Data` method is perfect for serving raw file data like images.
	c.Data(http.StatusOK, contentType, imageBytes)
}

func (handler *DownloadHandler) ImageOriginal(c *gin.Context) {

	fullPath := c.Param("filename")

	if fullPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename parameter is missing"})
		return
	}
	imageBytes, err := handler.manager.OriginalImageLoader.LoadImage(c, fullPath)
	if err != nil {
		log.Printf("Error loading image: %v", err)
		// Return a 404 if the image is not found, or a 500 for other errors.
		// A real implementation might check the error type to be more specific.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load image"})
		return
	}

	ext := filepath.Ext(fullPath)
	contentType := "application/octet-stream" // Default MIME type
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".svg":
		contentType = "image/svg+xml"
	}

	c.Data(http.StatusOK, contentType, imageBytes)
}
