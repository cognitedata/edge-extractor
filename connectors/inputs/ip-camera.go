package inputs

import (
	"fmt"

	"github.com/cognitedata/edge-extractor/drivers/camera"
)

type IpCamera struct {
	model    string
	address  string
	username string
	password string
	cType    string
	driver   camera.Driver
}

func NewIpCamera(model, address, cType, username, password string) *IpCamera {
	driverCon := map[string]camera.DriverConstructor{
		"fscam":      camera.NewFileSystemCameraDriver,
		"axis":       camera.NewAxisCameraDriver,
		"hickvision": camera.NewHikvisionCameraDriver,
		"reolink":    camera.NewReolinkCameraDriver,
	}

	driver := driverCon[model]

	if driver == nil {
		return nil
	}

	c := IpCamera{model: model, address: address, cType: cType, driver: driver(), username: username, password: password}
	return &c
}

func (cam *IpCamera) ExtractImage() (*camera.Image, error) {
	if cam.driver == nil {
		return nil, fmt.Errorf("unknown driver")
	}
	return cam.driver.ExtractImage(cam.address, cam.username, cam.password)
}
