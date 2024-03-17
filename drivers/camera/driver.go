package camera

type Image struct {
	Body          []byte
	Format        string
	TransactionId string
	ExternalId    string
}

type CameraEvent struct {
	Type string
	Data []byte
}

type EventFilter struct {
	TopicFilter   string
	ContentFilter string
}

type CameraManifest struct {
	Make                         string
	Model                        string
	IsCameraEventStreamSupported bool
}

type DriverConstructor func() Driver

type Driver interface {
	Configure(address, username, password string) error
	ExtractImage() (*Image, error)
	ExtractMetadata() ([]byte, error)
	Ping(address string) bool
	Commit(transactionId string) error
	SubscribeToEventsStream(eventFilters []EventFilter) (chan CameraEvent, error)
}
