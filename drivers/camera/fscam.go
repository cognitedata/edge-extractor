package camera

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"sync"
)

type FileSystemCameraDriver struct {
	fileCursor int
	dirContent []fs.DirEntry
	cursorMux  sync.Mutex
}

func NewFileSystemCameraDriver() Driver {
	return &FileSystemCameraDriver{cursorMux: sync.Mutex{}}
}

// ExtractImage reads the file from the file system and returns the image. If address is a directory, it will read the files in the directory using a cursor.
func (cam *FileSystemCameraDriver) ExtractImage(address, username, password string) (*Image, error) {

	if cam.dirContent != nil {
		// saved cursor is not nil, so we have already read the directory
		var img *Image
		var err error
		cam.cursorMux.Lock()
		if cam.fileCursor >= len(cam.dirContent) {
			cam.fileCursor = 0
			cam.dirContent = nil
		} else {
			fullPath := path.Join(address, cam.dirContent[cam.fileCursor].Name())
			img, err = cam.processFile(fullPath)
			img.ExternalId = strconv.Itoa(cam.fileCursor)
			cam.fileCursor++
		}
		cam.cursorMux.Unlock()
		return img, err

	} else {
		// readiung the directory for the first time and saving into memmory or processing single file
		file, err := os.Stat(address)
		if err != nil {
			return nil, err
		}
		mode := file.Mode()
		if mode.IsDir() {
			cam.dirContent, err = os.ReadDir(address)
			if err != nil {
				return nil, err
			}
			if len(cam.dirContent) == 0 {
				cam.dirContent = nil
				cam.fileCursor = 0
				return nil, nil
			}
			return cam.ExtractImage(address, username, password)
		} else {
			return cam.processFile(address)
		}
	}
}

// processFile reads the file , reads binary content into memory and returns the image
func (cam *FileSystemCameraDriver) processFile(filePath string) (*Image, error) {
	body, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	img := Image{Body: body, Format: "image/jpeg", TransactionId: filePath, ExternalId: "0"}

	return &img, nil
}

func (cam *FileSystemCameraDriver) ExtractMetadata(address, username, password string) ([]byte, error) {
	return nil, nil
}

func (cam *FileSystemCameraDriver) Ping(address string) bool {
	return true
}

// Commit removes the file from the file system. TransactionId in this case is the file path.
func (cam *FileSystemCameraDriver) Commit(transactionId string) error {
	os.Remove(transactionId)
	return nil
}

func (cam *FileSystemCameraDriver) SubscribeToEventsStream(address, username, password string) (chan CameraEvent, error) {
	return nil, fmt.Errorf("file system camera driver does not support event streaming")
}
