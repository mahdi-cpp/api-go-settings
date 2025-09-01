package thumbnail

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cshum/vipsgen/vips"
)

const userID = "018f3a8b-1b32-729a-f7e5-5467c1b2d3e4"
const workers = 1
const maxDimension = 270
const targetWidth = 270

func isImageFile(entry os.DirEntry) bool {
	name := strings.ToLower(entry.Name())
	return !entry.IsDir() &&
		(strings.HasSuffix(name, ".jpg") ||
			strings.HasSuffix(name, ".jpeg") ||
			strings.HasSuffix(name, ".heic") ||
			strings.HasSuffix(name, ".png"))
}

func CreateSingleThumbnail(src string, fileName string) error {

	_, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := processImage(src, "app/tmp/ali/"+fileName); err != nil {
		log.Printf("failed create single thumbnail %s: %v", src, err)
	}

	return nil
}

func CreateThumbnails() error {

	basePath := filepath.Join("/app/iris/com.iris.photos/users", userID, "zz")
	thumbPath := filepath.Join(basePath, "thumbnails")

	if err := os.MkdirAll(thumbPath, 0755); err != nil {
		return fmt.Errorf("failed to create thumbnails directory: %w", err)
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	jobs := make(chan string, len(entries))
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range jobs {
				fileName := filepath.Base(filePath)
				dest := filepath.Join(thumbPath, fileName)
				if err := processImage(filePath, dest); err != nil {
					log.Printf("Error processing %s: %v", fileName, err)
				}
			}
		}()
	}

	for _, entry := range entries {
		if !isImageFile(entry) {
			continue
		}
		src := filepath.Join(basePath, entry.Name())
		jobs <- src
	}
	close(jobs)

	wg.Wait()
	return nil
}

func processImage(filePath string, savePath string) error {

	// First get img dimensions to determine orientation
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	source := vips.NewSource(file)
	img, err := vips.NewImageFromSource(source, nil)
	if err != nil {
		return fmt.Errorf("failed to get img dimensions: %w", err)
	}
	defer img.Close()
	defer source.Close()

	fmt.Println("img orientation ", img.Orientation())
	var width = 0
	var height = 0
	if img.Orientation() == 6 {
		width = img.Height()
		height = img.Width()
	} else {
		width = img.Width()
		height = img.Height()
	}

	// Calculate resize scale based on orientation
	var scale float64
	if width >= height { // Landscape or square
		scale = float64(targetWidth) / float64(width)
	} else { // Portrait
		scale = float64(targetWidth) / float64(width)
	}

	// Resize the img
	err = img.Resize(scale, &vips.ResizeOptions{Kernel: vips.KernelNearest})
	if err != nil {
		fmt.Println("1", err.Error())
		return fmt.Errorf("failed to resize img: %w", err)
	}

	//img.Close()

	// Reopen the file for thumbnail processing
	//file2, err := os.Open(filePath)
	//if err != nil {
	//	return fmt.Errorf("failed to reopen file: %w", err)
	//}
	//defer file2.Close()

	//source2 := vips.NewSource(img)
	//defer source2.Close()

	//var thumbImage *vips.Image

	//if img.Format() == vips.ImageTypeHeif {
	//err = img.Heifsave(savePath, &vips.HeifsaveOptions{})
	//if err != nil {
	//	fmt.Println("2 error Heifsave ", err.Error())
	//	return err
	//}
	//} else {

	// Check if the filename has a .heic extension.
	if strings.HasSuffix(savePath, ".heic") {
		// Replace the extension with .jpg.
		savePath = strings.TrimSuffix(savePath, filepath.Ext(savePath)) + ".jpg"
	}

	err = img.Jpegsave(savePath, &vips.JpegsaveOptions{})
	if err != nil {
		fmt.Println("2 error Jpegsave ", err.Error())
		return err
	}
	//}

	//source.Close()

	//if width >= height { // Landscape or square
	// For landscape, set width to maxDimension, height will be calculated automatically
	//thumbImage, err = vips.NewThumbnailBuffer(img.GetB, maxDimension, &vips.ThumbnailSourceOptions{
	//	Height: 0, // Let vips calculate height automatically
	//})

	//a, err = vips.REs
	//} else { // Portrait
	//	// For portrait, we need to set both dimensions but use Crop to maintain aspect ratio
	//	thumbImage, err = vips.NewThumbnailSource(source2, maxDimension, &vips.ThumbnailSourceOptions{
	//		Height: maxDimension,
	//		Crop:   vips.InterestingCentre, // Crop to maintain aspect ratio
	//	})
	//}

	//if err != nil {
	//	return fmt.Errorf("failed to create thumbnail: %w", err)
	//}
	//defer thumbImage.Close()

	// Save with quality options
	//err = thumbImage.Jpegsave(savePath, &vips.JpegsaveOptions{
	//	//Quality: 85,
	//})
	//if err != nil {
	//	return fmt.Errorf("failed to save img: %w", err)
	//}

	log.Printf("Successfully created thumbnail for %s", filepath.Base(filePath))
	return nil
}

func main() {

	vips.Startup(nil)
	defer vips.Shutdown()

	start := time.Now()
	if err := CreateThumbnails(); err != nil {
		log.Fatalf("An error occurred during thumbnail creation: %v", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("Thumbnail creation took %s to run.\n", elapsed)
}
