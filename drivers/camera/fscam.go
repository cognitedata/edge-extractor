package camera

import (
	"io/ioutil"
	"os"
)

type FileSystemCameraDriver struct {
}

func NewFileSystemCameraDriver() Driver {
	return &FileSystemCameraDriver{}
}

func (cam *FileSystemCameraDriver) ExtractImage(address, username, password string) (*Image, error) {
	body, err := ioutil.ReadFile(address)
	if err != nil {
		return nil, err
	}

	os.Remove(address)

	img := Image{Body: body, Format: "image/jpeg"}

	return &img, nil
}

func (cam *FileSystemCameraDriver) ExtractMetadata(address, username, password string) ([]byte, error) {
	return nil, nil
}

func (cam *FileSystemCameraDriver) Ping(address string) bool {
	return true
}
