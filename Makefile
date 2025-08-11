default: build

build:
	@ echo "+ $@"
	@ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -C . -a -installsuffix '' -o ./bin/mdex ./cmd/mdex/

test:
	@ echo "+ $@"
	@ go test -C . -v ./...

build-linux:
	@ echo "+ $@"
	@ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ./bin/mdex-linux-amd64 ./cmd/mdex

build-darwin:
	@ echo "+ $@"
	@ CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -installsuffix cgo -o ./bin/mdex-darwin-amd64 ./cmd/mdex

build-windows:
	@ echo "+ $@"
	@ CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -o ./bin/mdex-windows-amd64 ./cmd/mdex
