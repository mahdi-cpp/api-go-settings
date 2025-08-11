package main

import (
	"github.com/mahdi-cpp/api-go-settings/internal/api/handler"
	"github.com/mahdi-cpp/api-go-settings/internal/storage"
	"log"
)

func main() {

	userStorageManager, err := storage.NewMainStorageManager()
	if err != nil {
		log.Fatal(err)
	}

	userHandler := handler.NewUserHandler(userStorageManager)
	routUserHandler(userHandler)

	assetHandler := handler.NewAssetHandler(userStorageManager)
	routAssetHandler(assetHandler)

	searchHandler := handler.NewSearchHandler(userStorageManager)
	routSearchHandler(searchHandler)

	albumHandler := handler.NewAlbumHandler(userStorageManager)
	routAlbumHandler(albumHandler)

	startServer(router)
}

func routUserHandler(userHandler *handler.UserHandler) {

	api := router.Group("/api/v1/user")

	api.POST("create", userHandler.Create)
	api.POST("update", userHandler.Update)
	api.POST("delete", userHandler.Delete)
	api.POST("user", userHandler.GetUserByID)
	api.POST("list", userHandler.GetList)
}

func routAssetHandler(assetHandler *handler.AssetHandler) {

	api := router.Group("/api/v1/assets")

	api.POST("create", assetHandler.Create)
	api.POST("assets", assetHandler.Upload)
	api.GET(":id", assetHandler.Get)
	api.POST("update", assetHandler.Update)
	api.POST("update_all", assetHandler.UpdateAll)
	api.POST("delete", assetHandler.Delete)
	api.POST("filters", assetHandler.Filters)

	api.GET("download/:filename", assetHandler.OriginalDownload)
	api.GET("download/thumbnail/:filename", assetHandler.TinyImageDownload)
	api.GET("download/icons/:filename", assetHandler.IconDownload)
}

func routSearchHandler(searchHandler *handler.SearchHandler) {
	api := router.Group("/api/v1/search")

	api.GET("", searchHandler.Search)
	api.POST("filters", searchHandler.Filters)
}

func routAlbumHandler(albumHandler *handler.AlbumHandler) {
	api := router.Group("/api/v1/album")

	api.POST("create", albumHandler.Create)
	api.POST("update", albumHandler.Update)
	api.POST("delete", albumHandler.Delete)
	api.POST("list", albumHandler.GetListV2)
}
