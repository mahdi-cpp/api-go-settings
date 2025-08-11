package storage

import (
	"context"
	"github.com/mahdi-cpp/api-go-pkg/asset_metadata_manager"
	"github.com/mahdi-cpp/api-go-pkg/collection"
	"github.com/mahdi-cpp/api-go-pkg/image_loader"
	"github.com/mahdi-cpp/api-go-pkg/shared_model"
	"github.com/mahdi-cpp/api-go-pkg/thumbnail"
	"github.com/mahdi-cpp/api-go-settings/internal/domain/model"
	"sync"
	"time"
)

type AppStorage struct {
	mu                  sync.RWMutex // Protects all indexes and maps
	user                shared_model.User
	originalImageLoader *image_loader.ImageLoader
	tinyImageLoader     *image_loader.ImageLoader
	assets              map[int]*shared_model.PHAsset
	AlbumManager        *collection.Manager[*model.Album]
	VillageManager      *collection.Manager[*model.Village]
	metadata            *asset_metadata_manager.AssetMetadataManager
	thumbnail           *thumbnail.ThumbnailManager
	lastID              int
	lastRebuild         time.Time
	maintenanceCtx      context.Context
	cancelMaintenance   context.CancelFunc
	statsMu             sync.Mutex
}
