package handler_v2

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-settings/internal/application"
)

type DownloadHandler struct {
	manager *application.AppManager
}

func NewDownloadHandler(manager *application.AppManager) *DownloadHandler {
	return &DownloadHandler{
		manager: manager,
	}
}

// serveImage handles common image serving logic
func (h *DownloadHandler) serveImage(c *gin.Context, loader func(context.Context, string) ([]byte, error)) {

	fullPath := c.Param("filename")
	if fullPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename parameter is missing"})
		return
	}

	imageBytes, err := loader(c, fullPath)
	if err != nil {
		log.Printf("Error loading image: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load image"})
		return
	}

	// Determine content type from file extension
	ext := filepath.Ext(fullPath)
	contentType := getContentType(ext)
	c.Data(http.StatusOK, contentType, imageBytes)
}

// getContentType returns the appropriate MIME type for file extensions
func getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// http://localhost:50000/api/v1/download/original
// -----------------------------------------------/com.iris.photos
// ---------------------------------------------------------------/users/018f3a8b-1b32-729a-f7e5-5467c1b2d3e4/assets/0198c111-0f9d-74f6-ab2e-6ce665ec29c6.jpg
// http://localhost:50000/api/v1/download/original/com.iris.photos/users/018f3a8b-1b32-729a-f7e5-5467c1b2d3e4/assets/0198c111-0f9d-74f6-ab2e-6ce665ec29c6.jpg

// ImageOriginal serves original images
func (h *DownloadHandler) ImageOriginal(c *gin.Context) {
	h.serveImage(c, h.manager.OriginalImageLoader.LoadImage)
}

// http://localhost:50000/api/v1/download/
// ---------------------------------------thumbnail
// ------------------------------------------------/com.iris.photos
// ----------------------------------------------------------------/users/018f3a8b-1b32-729a-f7e5-5467c1b2d3e4/assets/thumbnails/0198c111-0f9d-74f6-ab2e-6ce665ec29c6_270.jpg
// http://localhost:50000/api/v1/download/thumbnail/com.iris.photos/users/018f3a8b-1b32-729a-f7e5-5467c1b2d3e4/assets/thumbnails/0198c111-0f9d-74f6-ab2e-6ce665ec29c6_270.jpg

// ImageThumbnail serves thumbnail images
func (h *DownloadHandler) ImageThumbnail(c *gin.Context) {
	h.serveImage(c, h.manager.ThumbnailImageLoader.LoadImage)
}

// http://localhost:50000/api/v1/download/icon
// -------------------------------------------/com.iris.photos
// -----------------------------------------------------------/res/drawable/icons8-keyboard-100.png
// http://localhost:50000/api/v1/download/icon/com.iris.photos/res/drawable/icons8-keyboard-100.png

// ImageIcons serves icon images
func (h *DownloadHandler) ImageIcons(c *gin.Context) {
	h.serveImage(c, h.manager.IconImageLoader.LoadImage)
}
