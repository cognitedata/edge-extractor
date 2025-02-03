module github.com/cognitedata/edge-extractor

go 1.21

require (
	github.com/cognitedata/cognite-sdk-go v0.3.2-0.20211022150037-c6aa1283f946
	github.com/sirupsen/logrus v1.6.0
)

require github.com/cskr/pubsub/v2 v2.0.1 // indirect

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/gorilla/websocket v1.5.1
	github.com/kardianos/service v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/xinsnake/go-http-digest-auth-client v0.6.0
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3 // indirect
	golang.org/x/sys v0.28.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
)

replace github.com/cognitedata/cognite-sdk-go => ../../cognite-sdk-go
