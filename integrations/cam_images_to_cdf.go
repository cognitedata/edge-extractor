package integrations

import (
	"fmt"
	"time"

	"github.com/cognitedata/cognite-sdk-go/pkg/cognite/dto/core"
	"github.com/cognitedata/edge-extractor/connectors/inputs"
	"github.com/cognitedata/edge-extractor/internal"
	log "github.com/sirupsen/logrus"
)

type CameraImagesToCdfConfig struct {
}

type CameraImagesToCdf struct {
	cogClient *internal.CdfClient
	isStarted bool
}

func NewCameraImagesToCdf(cogClient *internal.CdfClient) *CameraImagesToCdf {
	return &CameraImagesToCdf{cogClient: cogClient}
}

// Introduce reload command.

func (intgr *CameraImagesToCdf) Start() error {
	intgr.isStarted = true

	filter := core.AssetFilter{Metadata: map[string]string{"cog_class": "camera", "state": "enabled"}}

	assets, err := intgr.cogClient.Client().Assets.Filter(filter, 1000)

	if err != nil {
		return err
	}

	// Creating processor for each camera asset.
	for ai := range assets {
		go intgr.startProcessor(20, &assets[ai])
	}
	return nil
}

// LoadConfigs load both static and dynamic configs
func (intgr *CameraImagesToCdf) LoadConfigs() {

}

// Start camera processor.
func (intgr *CameraImagesToCdf) startProcessor(delay int64, asset *core.Asset) error {
	log.Info("Starting camera processor ")
	// TODO : Investigate memmory usage with many cameras and big images . Next step is to introduce image streaming through file system to avoid excessive RAM usag.
	model := asset.Metadata["cog_model"]
	address := asset.Metadata["uri"]
	username := asset.Metadata["username"]
	password := asset.Metadata["password"]
	log.Infof(" Camera name = %s , model = %s , address = %s , username = %s", asset.Name, model, address, username)

	if model == "" || address == "" {
		log.Errorf("Processor can't be started for camera %s . Model or address aren't set.", asset.Name)
		return fmt.Errorf("empty asset model or address")
	}

	cam := inputs.NewIpCamera(model, address, "", username, password)
	if cam == nil {
		log.Error("Unsupported camera model")
		return fmt.Errorf("unsupported camera model")
	}
	for {
		img, err := cam.ExtractImage()
		if err != nil {
			log.Debug("Can't extract image. Error : ", err.Error())
			time.Sleep(time.Second * 30)
		} else {
			err := intgr.cogClient.UploadInMemoryFile(img.Body, "", asset.Name, img.Format, asset.ID)
			if err != nil {
				log.Debug("Can't upload image. Error : ", err.Error())
				time.Sleep(time.Second * 30)
			} else {
				log.Debug("File uploaded to CDF successfully")
			}
		}
		if !intgr.isStarted {
			break
		}
		time.Sleep(time.Second * time.Duration(delay))
	}

	log.Info("Processor exited main loop ")
	return nil

}
