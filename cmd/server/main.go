package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-settings/internal/api/handler"
	"github.com/mahdi-cpp/api-go-settings/internal/application"
	borker "github.com/mahdi-cpp/api-go-settings/internal/broker"
)

func main() {

	go borker.StartRespBroker()

	// Load HTML templates
	router.LoadHTMLGlob("/app/tmp/templates/*")

	// Create upload handler
	uploadHandler := &handler.UploadHandler{
		UploadDir: "/app/tmp/uploads",
	}
	// Setup routes
	setupRoutes(router, uploadHandler)

	newAppManager, err := application.NewAppManager()
	if err != nil {
		log.Fatal(err)
	}

	downloadHandler := handler.NewDownloadHandler(newAppManager)
	routDownloadHandler(downloadHandler)

	startServer(router)
}

func setupRoutes(router *gin.Engine, uploadHandler *handler.UploadHandler) {
	// Serve upload form
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// Setup upload routes
	routUploadHandler(router, uploadHandler)
}

func routUploadHandler(router *gin.Engine, uploadHandler *handler.UploadHandler) {
	api := router.Group("/api/v1/upload")

	api.POST("/jpeg", uploadHandler.UploadJPEG)
	api.POST("/multiple", uploadHandler.UploadMultiple)
	api.GET("/files", uploadHandler.ListFiles)
}

func routDownloadHandler(userHandler *handler.DownloadHandler) {

	api := router.Group("/api/v1/download")

	api.GET("original/*filename", userHandler.ImageOriginal)
	api.GET("thumbnail/*filename", userHandler.ImageThumbnail)
	api.GET("icon/*filename", userHandler.ImageIcons)
}
