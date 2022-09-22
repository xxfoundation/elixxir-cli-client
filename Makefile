.PHONY: update master release update_master update_release build clean version

version:
	go run main.go generate
	mv version_vars.go cmd/version_vars.go

clean:
	rm -rf vendor/
	go mod vendor -e

update:
	-GOFLAGS="" go get all

build:
	go build ./...
	go mod tidy

update_project:
	GOFLAGS="" go get -d gitlab.com/elixxir/client@project/Channels
	GOFLAGS="" go get -d gitlab.com/elixxir/crypto@project/channels
	GOFLAGS="" go get -d gitlab.com/elixxir/primitives@release
	GOFLAGS="" go get -d gitlab.com/xx_network/crypto@release
	GOFLAGS="" go get -d gitlab.com/xx_network/primitives@project/channels

update_release:
	GOFLAGS="" go get -d gitlab.com/elixxir/client@release
	GOFLAGS="" go get -d gitlab.com/elixxir/crypto@release
	GOFLAGS="" go get -d gitlab.com/elixxir/primitives@release
	GOFLAGS="" go get -d gitlab.com/xx_network/crypto@release
	GOFLAGS="" go get -d gitlab.com/xx_network/primitives@release

update_master:
	GOFLAGS="" go get -d gitlab.com/elixxir/client@master
	GOFLAGS="" go get -d gitlab.com/elixxir/crypto@master
	GOFLAGS="" go get -d gitlab.com/elixxir/primitives@master
	GOFLAGS="" go get -d gitlab.com/xx_network/crypto@master
	GOFLAGS="" go get -d gitlab.com/xx_network/primitives@master

master: update_master clean build version

release: update_release clean build

project: update_project clean build
