package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-pkg/common_models"
	"github.com/mahdi-cpp/api-go-settings/internal/storage"
	"net/http"
)

type UserHandler struct {
	settingStorageManager *storage.SettingStorageManager
}

func NewUserHandler(userStorageManager *storage.SettingStorageManager) *UserHandler {
	return &UserHandler{
		settingStorageManager: userStorageManager,
	}
}

func (handler *UserHandler) Create(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var request common_models.UserHandler
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userStorage, err := handler.settingStorageManager.GetUserStorage(c, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	newItem, err := handler.settingStorageManager.UserManager.Create(&common_models.User{
		Username:    request.Username,
		FirstName:   request.FirstName,
		LastName:    request.LastName,
		AvatarURL:   request.AvatarURL,
		PhoneNumber: request.PhoneNumber,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	//update := common_models.AssetUpdate{AssetIds: request.AssetIds, AddAlbums: []int{newItem.ID}}
	//_, err = userStorage.UpdateAsset(update)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	//	return
	//}

	userStorage.UpdateCollections()

	c.JSON(http.StatusCreated, common_models.CollectionResponse{
		ID:    newItem.ID,
		Title: newItem.Username,
	})
}

func (handler *UserHandler) Update(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var itemHandler common_models.UserHandler
	if err := c.ShouldBindJSON(&itemHandler); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	item, err := handler.settingStorageManager.UserManager.Get(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	common_models.UpdateUser(item, itemHandler)

	item2, err := handler.settingStorageManager.UserManager.Update(item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *UserHandler) Delete(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	var item common_models.User
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err = handler.settingStorageManager.UserManager.Delete(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, "delete ok")
}

func (handler *UserHandler) GetCollectionList(c *gin.Context) {

	item2, err := handler.settingStorageManager.UserManager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusCreated, item2)
}

func (handler *UserHandler) GetUserByID(c *gin.Context) {

	userID, err := getUserId(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "userID must be an integer"})
		return
	}

	user, err := handler.settingStorageManager.UserManager.Get(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	//result := common_models.PHCollectionList[*common_models.User]{
	//	Collections: make([]*common_models.PHCollection[*common_models.User], len(items)),
	//}
	//
	//for i, item := range items {
	//	assets, _ := handler.settingStorageManager.UserManager.GetItemAssets(item.ID)
	//	result.Collections[i] = &common_models.PHCollection[*common_models.User]{
	//		Item:   item,
	//		Assets: assets,
	//	}
	//}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (handler *UserHandler) GetList(c *gin.Context) {

	//var with common_models.PHFetchOptions
	//if err := c.ShouldBindJSON(&with); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
	//	fmt.Println("Invalid request")
	//	return
	//}

	//userStorage, err := handler.settingStorageManager.GetUserStorage(c, userID)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err})
	//}

	//items, err := handler.settingStorageManager.UserManager.GetAllSorted(with.SortBy, with.SortOrder)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": err})
	//	return
	//}

	items, err := handler.settingStorageManager.UserManager.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	result := common_models.PHCollectionList[*common_models.User]{
		Collections: make([]*common_models.PHCollection[*common_models.User], len(items)),
	}

	for i, item := range items {
		assets, _ := handler.settingStorageManager.UserManager.GetItemAssets(item.ID)
		result.Collections[i] = &common_models.PHCollection[*common_models.User]{
			Item:   item,
			Assets: assets,
		}
	}

	c.JSON(http.StatusOK, result)
}
