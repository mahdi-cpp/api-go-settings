package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-settings/internal/api/handler"
	"github.com/mahdi-cpp/api-go-settings/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	userStorageManager, err := storage.NewSettingStorageManager()
	if err != nil {
		log.Fatal(err)
	}

	userHandler := handler.NewUserHandler(userStorageManager)
	assetHandler := handler.NewAssetHandler(userStorageManager)
	searchHandler := handler.NewSearchHandler(userStorageManager)
	albumHandler := handler.NewAlbumHandler(userStorageManager)

	// Handler Gin router
	router := createRouter(
		userHandler,
		assetHandler,
		albumHandler,
		searchHandler,
	)

	// Start server
	startServer(router)
}

func createRouter(
	userHandler *handler.UserHandler,
	assetHandler *handler.AssetHandler,
	albumHandler *handler.AlbumHandler,
	searchHandler *handler.SearchHandler,
) *gin.Engine {

	// Set Gin mode
	gin.SetMode("release")

	// Handler router with default middleware
	router := gin.Default()

	// API routes
	api := router.Group("/api/v1")
	{

		api.POST("/user/create", userHandler.Create)
		api.POST("/user/update", userHandler.Update)
		api.POST("/user/delete", userHandler.Delete)
		api.POST("/user/user", userHandler.GetUserByID)
		api.POST("/user/list", userHandler.GetList)

		// Search routes
		api.GET("/search", searchHandler.Search)
		api.POST("/search/filters", searchHandler.Filters)

		// Asset routes
		api.POST("/assets/create", assetHandler.Create)
		api.POST("/assets", assetHandler.Upload)
		api.GET("/assets/:id", assetHandler.Get)
		api.POST("/assets/update", assetHandler.Update)
		api.POST("/assets/update_all", assetHandler.UpdateAll)
		api.POST("/assets/delete", assetHandler.Delete)
		api.POST("/assets/filters", assetHandler.Filters)

		//http://localhost:8080/api/v1/assets/download/thumbnail/map_270.jpg
		api.GET("/assets/download/:filename", assetHandler.OriginalDownload)
		api.GET("/assets/download/thumbnail/:filename", assetHandler.TinyImageDownload)
		api.GET("/assets/download/icons/:filename", assetHandler.IconDownload)

		api.POST("/album/create", albumHandler.Create)
		api.POST("/album/update", albumHandler.Update)
		api.POST("/album/delete", albumHandler.Delete)
		api.POST("/album/list", albumHandler.GetListV2)

	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func startServer(router *gin.Engine) {

	// Handler HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "0.0.0.0", 8080),
		Handler: router,
	}

	// Run server in a goroutine
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
