package cam_images_to_cdf

type CameraConfig struct {
	ID              uint64
	ExternalID      string
	Name            string
	Model           string
	Address         string
	Username        string
	Password        string
	Mode            string
	PollingInterval int
	State           string
	LinkedAssetID   uint64
	IsEncrypted     bool
}

// Compare CameraConfig with anothert CameraConfig
func (c *CameraConfig) IsEqual(other *CameraConfig) bool {
	return c.Name == other.Name &&
		c.Model == other.Model &&
		c.Address == other.Address &&
		c.Username == other.Username &&
		c.Password == other.Password &&
		c.Mode == other.Mode &&
		c.PollingInterval == other.PollingInterval &&
		c.State == other.State &&
		c.LinkedAssetID == other.LinkedAssetID
}

type CameraImagesToCdfConfig struct {
	Cameras []CameraConfig
}
