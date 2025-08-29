package main

import (
	"log"

	"github.com/mahdi-cpp/api-go-settings/internal/api/handler_v2"
	"github.com/mahdi-cpp/api-go-settings/internal/application"
	borker "github.com/mahdi-cpp/api-go-settings/internal/broker"
)

func main() {

	go borker.StartRespBroker()

	newAppManager, err := application.NewAppManager()
	if err != nil {
		log.Fatal(err)
	}

	downloadHandler := handler_v2.NewDownloadHandler(newAppManager)
	routDownloadHandler(downloadHandler)

	startServer(router)
}

func routDownloadHandler(userHandler *handler_v2.DownloadHandler) {

	api := router.Group("/api/v1/download")

	api.GET("original/*filename", userHandler.ImageOriginal)
	api.GET("thumbnail/*filename", userHandler.ImageThumbnail)
	api.GET("icon/*filename", userHandler.ImageIcons)
}
