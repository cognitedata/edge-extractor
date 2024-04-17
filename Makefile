version="0.7.1"
encryption_key=${CONFIG_ENCRYPTION_KEY}
test_encryption_key=test_key_ZU8uJ8vxJs7Z75uF3Jy8m52 # gitleaks:allow

run :
	go run cmd/main.go --config config_encrypted_local.json --key ${test_encryption_key}

run-encrypt-config :
	go run cmd/main.go --config config_plain_local.json --key ${test_encryption_key} --op encrypt_config

build : 
	go build -o edge-extractor -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-win-amd64.exe cmd/main.go

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-win-arm64.exe cmd/main.go


build-linux-386:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-linux-386 cmd/main.go

build-linux-arm:
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-linux-arm cmd/main.go

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-linux-amd64 cmd/main.go

build-osx-intel:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-osx-amd64 cmd/main.go

build-osx-arm:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.Version=${version} -X main.EncryptionKey=${encryption_key}" -o edge-extractor-osx-arm cmd/main.go


build-multios: build-windows build-windows-arm64 build-linux-386 build-linux-amd64 build-linux-arm build-osx-intel build-osx-arm

build-docker:
	docker build -t cognite/edge-extractor:${version} -t cognite/edge-extractor:latest .

test-docker:
	docker run -it --rm --name edge-extractor cognite/edge-extractor:${version} 
prepare-test-data:
	cp imgdump-src/* imgdump/

set_private_repos :
	go get github.com/cognitedata/cognite-sdk-go

copy-binary-to-remote:
	scp edge-extractor-linux-amd64 ubuntu@192.168.202.10:/home/ubuntu/edge-extractor
