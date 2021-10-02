
set_private_repos :
	go get github.com/cognitedata/cognite-sdk-go

build : 
	go build -o edge-extractor cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o edge-extractor.exe cmd/main.go

build-386:
	GOOS=linux GOARCH=386 go build -o edge-extractor cmd/main.go


