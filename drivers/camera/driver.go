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

type CameraManifest struct {
	Make                         string
	Model                        string
	IsCameraEventStreamSupported bool
}

type DriverConstructor func() Driver

type Driver interface {
	ExtractImage(address, username, password string) (*Image, error)
	ExtractMetadata(address, username, password string) ([]byte, error)
	Ping(address string) bool
	Commit(transactionId string) error
	SubscribeToEventsStream(address, username, password string) (chan CameraEvent, error)
}
