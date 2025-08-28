package application

import (
	"sync"

	"github.com/mahdi-cpp/api-go-pkg/image_loader"
	"github.com/mahdi-cpp/api-go-settings/config"
)

type AppManager struct {
	mu                   sync.RWMutex
	IconImageLoader      *image_loader.ImageLoader
	OriginalImageLoader  *image_loader.ImageLoader
	ThumbnailImageLoader *image_loader.ImageLoader
}

func NewAppManager() (*AppManager, error) {

	manager := &AppManager{}

	manager.IconImageLoader = image_loader.NewImageLoader(5000, config.GetRootDir(), 0)
	manager.OriginalImageLoader = image_loader.NewImageLoader(100, config.GetRootDir(), 0)
	manager.ThumbnailImageLoader = image_loader.NewImageLoader(5000, config.GetRootDir(), 0)

	return manager, nil
}
