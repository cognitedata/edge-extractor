package camera

type Image struct {
	Body          []byte
	Format        string
	TransactionId string
	ExternalId    string
}

type CameraEvent struct {
	CoreType  string
	Type      string
	Topic     string
	Source    string
	Timestamp int64
	RawData   []byte
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

type CameraServiceDiscoveryManifest struct {
	Name          string
	Format        string // json, xml, soap
	ComponentName string // Component name , for example "services" , "events", etc.
	Body          []byte
}

type DriverConstructor func() Driver

type Driver interface {
	Configure(address, username, password string) error
	ExtractImage() (*Image, error)
	ExtractMetadata() ([]byte, error)
	Ping(address string) bool
	Commit(transactionId string) error
	SubscribeToEventsStream(eventFilters []EventFilter) (chan CameraEvent, error)
	GetServicesDiscoveryManifest(componentName string) ([]CameraServiceDiscoveryManifest, error)
	Close()
}
