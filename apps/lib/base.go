package lib

import "encoding/json"

type AppInstance interface {
	// ConfigureFromRaw configures the app instance using the provided raw json configuration data.
	// It returns an error if the configuration is invalid or cannot be parsed.
	ConfigureFromRaw(config json.RawMessage) error

	// ConfigureIntegration configures the app instance with the provided integration.
	ConfigureIntegration(integration interface{})

	// GetDependencies returns the dependencies required by the app instance.
	GetDependencies() AppDependencies

	// Start starts the app instance.
	// It returns an error if the app fails to start.
	Start() error

	// Stop stops the app instance.
	// It returns an error if the app fails to stop.
	Stop() error
}

type AppDependencies struct {
	Integrations []string
}

type AppInstanceConstructor func() AppInstance

func NewAppInstance(name string) AppInstance {
	appInstance := map[string]AppInstanceConstructor{
		"CameraEventBasedCaptureApp": NewCameraEventBasedCaptureApp,
	}

	if constructor, ok := appInstance[name]; ok {
		return constructor()
	}
	return nil
}
