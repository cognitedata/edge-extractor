package inputs

import (
	"fmt"

	"github.com/cognitedata/edge-extractor/drivers/camera"
)

type IpCamera struct {
	ID       uint64
	Name     string
	model    string
	address  string
	username string
	password string
	cType    string
	driver   camera.Driver
}

func NewIpCamera(ID uint64, name, model, address, cType, username, password string) *IpCamera {
	driverCon := map[string]camera.DriverConstructor{
		"fscam":     camera.NewFileSystemCameraDriver,
		"axis":      camera.NewAxisCameraDriver,
		"hikvision": camera.NewHikvisionCameraDriver,
		"reolink":   camera.NewReolinkCameraDriver,
		"urlcam":    camera.NewUrlCameraDriver,
		"flir_ax8":  camera.NewFlirAx8CameraDriver,
		"dahua":     camera.NewDahuaCameraDriver,
	}

	driver := driverCon[model]

	if driver == nil {
		return nil
	}

	c := IpCamera{ID: ID, Name: name, model: model, address: address, cType: cType, driver: driver(), username: username, password: password}
	c.Configure()
	return &c
}

func (cam *IpCamera) Configure() error {
	if cam.driver == nil {
		return fmt.Errorf("unknown driver")
	}
	return cam.driver.Configure(cam.address, cam.username, cam.password)
}

func (cam *IpCamera) ExtractImage() (*camera.Image, error) {
	if cam.driver == nil {
		return nil, fmt.Errorf("unknown driver")
	}
	return cam.driver.ExtractImage()
}

func (cam *IpCamera) SubscribeToEventsStream(eventFilters []camera.EventFilter) (chan camera.CameraEvent, error) {
	if cam.driver == nil {
		return nil, fmt.Errorf("unknown driver")
	}
	return cam.driver.SubscribeToEventsStream(eventFilters)
}

func (cam *IpCamera) ExtractMetadata() ([]byte, error) {
	if cam.driver == nil {
		return nil, fmt.Errorf("unknown driver")
	}
	return cam.driver.ExtractMetadata()
}

func (cam *IpCamera) Commit(transactionId string) error {
	if cam.driver == nil {
		return fmt.Errorf("unknown driver")
	}
	return cam.driver.Commit(transactionId)
}

func (cam *IpCamera) GetDriver() camera.Driver {
	return cam.driver
}

func (cam *IpCamera) Close() {
	if cam.driver != nil {
		cam.driver.Close()
	}
}

func (cam *IpCamera) GetCameraCapabilitiesManifest(componentName string) ([]camera.CameraCapabilitiesManifest, error) {
	if cam.driver == nil {
		return nil, fmt.Errorf("unknown driver")
	}
	return cam.driver.GetCameraCapabilitiesManifest(componentName)
}
