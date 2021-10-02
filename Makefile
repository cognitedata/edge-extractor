
set_private_repos :
	go get github.com/cognitedata/cognite-sdk-go

build : 
	go build -o edge-extractor cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o edge-extractor-win-amd64.exe cmd/main.go

build-linux-386:
	GOOS=linux GOARCH=386 go build -o edge-extractor-linux-386 cmd/main.go

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o edge-extractor-linux-386 cmd/main.go

build-osx-intel:
	GOOS=darwin GOARCH=amd64 go build -o edge-extractor-osx-amd64 cmd/main.go

build-osx-arm:
	GOOS=darwin GOARCH=arm64 go build -o edge-extractor-osx-arm cmd/main.go


build-multios: build-windows build-linux-386 build-linux-amd64 build-osx-intel build-osx-arm


