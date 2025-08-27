package handler

import "C"
import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-pkg/account"
	"github.com/mahdi-cpp/api-go-pkg/collection"
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

func (handler *DownloadHandler) Create(c *gin.Context) {

	var request collection.CollectionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userManager, err := handler.manager.GetUserManager(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	newItem, err := userManager.GetCollections().Album.Collection.Create(&album.Album{Title: request.Title})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	update := phasset.UpdateOptions{
		AssetIds:  request.AssetIds,
		AddAlbums: []string{newItem.ID},
	}
	_, err = userManager.UpdateAssets(update)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userManager.UpdateCollections()

	c.JSON(http.StatusCreated, CollectionResponse{
		ID:    newItem.ID,
		Title: newItem.Title,
	})
}
