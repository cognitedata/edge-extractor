package ip_cams_to_cdf

type CameraConfig struct {
	ID                      uint64
	ExternalID              string
	Name                    string
	Model                   string
	Address                 string
	Username                string
	Password                string
	Mode                    string
	PollingInterval         int
	State                   string
	LinkedAssetID           uint64
	EnableCameraEventStream bool
	EventFilters            []CameraEventFilter
}

type CameraEventFilter struct {
	TopicFilter   string
	ContentFilter string
}

// Compare CameraConfig with anothert CameraConfig
func (c *CameraConfig) IsEqual(other *CameraConfig) bool {
	var isEventFiltersEqual bool
	if len(c.EventFilters) != len(other.EventFilters) {
		return false
	}
	for i, eventFilter := range c.EventFilters {
		if eventFilter != other.EventFilters[i] {
			return false
		}
	}

	return c.Name == other.Name &&
		c.Model == other.Model &&
		c.Address == other.Address &&
		c.Username == other.Username &&
		c.Password == other.Password &&
		c.Mode == other.Mode &&
		c.PollingInterval == other.PollingInterval &&
		c.State == other.State &&
		c.LinkedAssetID == other.LinkedAssetID &&
		c.EnableCameraEventStream == other.EnableCameraEventStream &&
		isEventFiltersEqual

}

type IntegrationConfig struct {
	Cameras             []CameraConfig
	DisableRunReporting bool
}

// Compare CameraImagesToCdfConfig with another CameraImagesToCdfConfig
func (c *IntegrationConfig) IsEqual(other *IntegrationConfig) bool {
	if len(c.Cameras) != len(other.Cameras) {
		return false
	}
	for i, camera := range c.Cameras {
		if !camera.IsEqual(&other.Cameras[i]) {
			return false
		}
	}
	return true
}

// clone returns a deep copy of IntegrationConfig
func (c *IntegrationConfig) Clone() IntegrationConfig {
	clone := IntegrationConfig{}
	clone.Cameras = make([]CameraConfig, len(c.Cameras))
	copy(clone.Cameras, c.Cameras)
	return clone
}
