package storage

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mahdi-cpp/api-go-pkg/collection"
	"github.com/mahdi-cpp/api-go-pkg/collection_controll"
	"github.com/mahdi-cpp/api-go-pkg/image_loader"
	"github.com/mahdi-cpp/api-go-pkg/metadata"
	"github.com/mahdi-cpp/api-go-pkg/shared_model"
	"github.com/mahdi-cpp/api-go-settings/config"
	"github.com/mahdi-cpp/api-go-settings/internal/domain/model"
	"sync"
	"time"
)

type MainStorageManager struct {
	mu           sync.RWMutex
	UserManager  *collection.Manager[*shared_model.User]
	usersStorage map[int]*UserStorage
	appsStorage  map[int]*AppStorage
	iconLoader   *image_loader.ImageLoader
	infoPlist    *metadata.Control[model.InfoPlist]
	ctx          context.Context
}

func NewMainStorageManager() (*MainStorageManager, error) {

	// Handler the storageManager
	storageManager := &MainStorageManager{
		usersStorage: make(map[int]*UserStorage),
		ctx:          context.Background(),
	}

	storageManager.infoPlist = metadata.NewMetadataControl[model.InfoPlist]("/media/mahdi/Cloud/Happle/com.helium.settings/Info.json")
	a, err := storageManager.infoPlist.Read(true)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("CFBundleDevelopmentRegion: ", a.CFBundleName)

	storageManager.UserManager, err = collection.NewCollectionManager[*shared_model.User](config.GetPath("data/users.json"), true)
	if err != nil {
		panic(err)
	}

	storageManager.iconLoader = image_loader.NewImageLoader(1000, config.GetPath("data/icons"), 0)
	storageManager.loadAllIcons()

	return storageManager, nil
}

func (us *MainStorageManager) GetUserStorage(c *gin.Context, userID int) (*UserStorage, error) {

	us.mu.Lock()
	defer us.mu.Unlock()

	var err error

	if userID <= 0 {
		return nil, fmt.Errorf("user id is Invalid")
	}

	user, err := us.UserManager.Get(userID)
	if err != nil {
		panic("user not found")
	}

	// Check if userStorage already exists for this user
	if storage, exists := us.usersStorage[userID]; exists {
		return storage, nil
	}

	fmt.Println("GetUserStorage 1-------------")
	// Handler context for background workers
	ctx, cancel := context.WithCancel(context.Background())

	// Ensure user directories exist
	//userDirs := []string{userAssetDir, userMetadataDir, userThumbnailsDir}
	//for _, dir := range userDirs {
	//	if err := os.MkdirAll(dir, 0755); err != nil {
	//		return nil, fmt.Errorf("failed to create user directory %s: %w", dir, err)
	//	}
	//}

	// Handler new userStorage for this user
	userStorage := &UserStorage{
		user: *user,
		//metadata:          metadata.NewMetadataManager(config.GetUserPath(user.PhoneNumber, "metadata")),
		//thumbnail:         thumbnail.NewThumbnailManager(config.GetUserPath("09355512619", "thumbnails")),
		maintenanceCtx:    ctx,
		cancelMaintenance: cancel,
	}

	userStorage.originalImageLoader = image_loader.NewImageLoader(50, config.GetUserPath(user.PhoneNumber, "assets"), 5*time.Minute)
	userStorage.tinyImageLoader = image_loader.NewImageLoader(30000, config.GetUserPath(user.PhoneNumber, "thumbnails"), 60*time.Minute)

	//userStorage.assets, err = userStorage.metadata.LoadUserAllMetadata()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to load metadata for user %s: %w", userID, err)
	//}

	userStorage.AlbumManager, err = collection_controll.NewCollectionManager[*model.Album](config.GetUserPath(user.PhoneNumber, "albums.json"), false)
	if err != nil {
		fmt.Println("UserStorage:", err)
	}

	userStorage.VillageManager, err = collection_controll.NewCollectionManager[*model.Village](config.GetPath("villages.json"), false)
	if err != nil {
		fmt.Println("UserStorage:", err)
	}

	userStorage.prepareAlbums()

	// Store the new userStorage
	us.usersStorage[userID] = userStorage

	return userStorage, nil
}

func (us *MainStorageManager) loadAllIcons() {
	us.iconLoader.GetLocalBasePath()

	// Scan metadata directory
	//files, err := os.ReadDir(us.iconLoader.GetLocalBasePath())
	//if err != nil {
	//	fmt.Println("failed to read metadata directory: %w", err)
	//}

	//var images []string
	//for _, file := range files {
	//	if strings.HasSuffix(file.Name(), ".png") {
	//		images = append(images, "/media/mahdi/Cloud/apps/Photos/parsa_nasiri/assets/"+file.Name())
	//	}
	//}
}

func (us *MainStorageManager) periodicMaintenance() {

	saveTicker := time.NewTicker(10 * time.Second)
	statsTicker := time.NewTicker(30 * time.Minute)
	rebuildTicker := time.NewTicker(24 * time.Hour)
	cleanupTicker := time.NewTicker(1 * time.Hour)

	for {
		select {
		case <-saveTicker.C:
			fmt.Println("saveTicker")
		case <-rebuildTicker.C:
			fmt.Println("rebuildTicker")
		case <-statsTicker.C:
			fmt.Println("statsTicker")
		case <-cleanupTicker.C:
			fmt.Println("cleanupTicker")
		}
	}
}

func (us *MainStorageManager) RepositoryGetOriginalImage(userID int, filename string) ([]byte, error) {
	return us.usersStorage[userID].originalImageLoader.LoadImage(us.ctx, filename)
}

func (us *MainStorageManager) RepositoryGetTinyImage(userID int, filename string) ([]byte, error) {
	a := us.usersStorage[userID]
	if a == nil {
		return nil, fmt.Errorf("tinyImageLoader is nill")
	} else {
		return us.usersStorage[userID].tinyImageLoader.LoadImage(us.ctx, filename)
	}
}

func (us *MainStorageManager) RepositoryGetIcon(filename string) ([]byte, error) {
	return us.iconLoader.LoadImage(us.ctx, filename)
}

func (us *MainStorageManager) RemoveStorageForUser(userID int) {
	us.mu.Lock()
	defer us.mu.Unlock()

	if storage, exists := us.usersStorage[userID]; exists {
		// Cancel any background operations
		storage.cancelMaintenance()
		// Remove from map
		delete(us.usersStorage, userID)
	}
}
