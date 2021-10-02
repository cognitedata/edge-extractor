package camera

type Image struct {
	Body   []byte
	Format string
}

type DriverConstructor func() Driver

type Driver interface {
	ExtractImage(address, username, password string) (*Image, error)
	Ping(address string) bool
}
