package application

import (
	"sync"

	"github.com/mahdi-cpp/api-go-pkg/image_loader"
	"github.com/mahdi-cpp/api-go-settings/config"
)

type AppManager struct {
	mu                   sync.RWMutex
	iconImageLoader      *image_loader.ImageLoader
	originalImageLoader  *image_loader.ImageLoader
	thumbnailImageLoader *image_loader.ImageLoader
}

func NewAppManager() (*AppManager, error) {

	manager := &AppManager{}

	manager.iconImageLoader = image_loader.NewImageLoader(5000, config.GetPath(""), 0)
	manager.originalImageLoader = image_loader.NewImageLoader(100, config.GetPath(""), 0)
	manager.thumbnailImageLoader = image_loader.NewImageLoader(5000, config.GetPath(""), 0)

	return manager, nil
}
