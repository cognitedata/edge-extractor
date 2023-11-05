version="0.4.0"
encryption_key=${CONFIG_ENCRYPTION_KEY}
set_private_repos :
	go get github.com/cognitedata/cognite-sdk-go

run :
	go run cmd/main.go

build : 
	go build -o edge-extractor -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-win-amd64.exe cmd/main.go

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-win-arm64.exe cmd/main.go


build-linux-386:
	GOOS=linux GOARCH=386 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-linux-386 cmd/main.go

build-linux-arm:
	GOOS=linux GOARCH=arm go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-linux-arm cmd/main.go

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-linux-amd64 cmd/main.go

build-osx-intel:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-osx-amd64 cmd/main.go

build-osx-arm:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-osx-arm cmd/main.go


build-multios: build-windows build-windows-arm64 build-linux-386 build-linux-amd64 build-linux-arm build-osx-intel build-osx-arm


prepare-test-data:
	cp imgdump-src/* imgdump/
