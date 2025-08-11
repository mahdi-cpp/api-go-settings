package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-pkg/shared_model"
	"github.com/mahdi-cpp/api-go-settings/internal/storage"
	"net/http"
)

type SearchHandler struct {
	userStorageManager *storage.MainStorageManager
}

func NewSearchHandler(userStorageManager *storage.MainStorageManager) *SearchHandler {
	return &SearchHandler{userStorageManager: userStorageManager}
}

func (handler *SearchHandler) Filters(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var with shared_model.PHFetchOptions
	if err := c.ShouldBindJSON(&with); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		fmt.Println("Invalid request")
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	items, total, err := userStorage.FetchAssets(with)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	fmt.Println("Filters count: ", len(items))

	result := shared_model.PHFetchResult[*shared_model.PHAsset]{
		Items:  items,
		Total:  total,
		Limit:  100,
		Offset: 100,
	}
	c.JSON(http.StatusOK, result)
}

func (handler *SearchHandler) Search(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var with shared_model.PHFetchOptions
	if err := c.ShouldBindJSON(&with); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.userStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	items, total, err := userStorage.FetchAssets(with)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	result := shared_model.PHFetchResult[*shared_model.PHAsset]{
		Items:  items,
		Total:  total,
		Limit:  100,
		Offset: 100,
	}
	c.JSON(http.StatusOK, result)
}
