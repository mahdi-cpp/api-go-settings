package main

import (
	"log"

	"github.com/mahdi-cpp/api-go-settings/internal/api/handler"
	"github.com/mahdi-cpp/api-go-settings/internal/application"
)

func main() {

	newAppManager, err := application.NewAppManager()
	if err != nil {
		log.Fatal(err)
	}

	downloadHandler := handler.NewDownloadHandler(newAppManager)
	routDownloadHandler(downloadHandler)

	startServer(router)
}

func routDownloadHandler(userHandler *handler.DownloadHandler) {

	api := router.Group("/api/v1/download_")

	api.GET("original/*filename", userHandler.ImageOriginal)
	api.GET("thumbnail/*filename", userHandler.ImageThumbnail)
}
