.PHONY: update master release update_master update_release build clean version

version:
	go run main.go generate
	sed -i.bak 's/package\ cmd/package\ api/g' version_vars.go
	mv version_vars.go api/version_vars.go

clean:
	rm -rf vendor/
	go mod vendor

update:
	-GOFLAGS="" go get all

build:
	go build ./...
	go mod tidy

update_release:
	GOFLAGS="" go get -d gitlab.com/xx_network/primitives@release
	GOFLAGS="" go get -d gitlab.com/xx_network/crypto@release
	GOFLAGS="" go get -d gitlab.com/elixxir/crypto@symmetricBroadcst
	GOFLAGS="" go get -d gitlab.com/elixxir/client@symmetricBroadcst

update_master:
	GOFLAGS="" go get gitlab.com/xx_network/crypto@master
	GOFLAGS="" go get gitlab.com/elixxir/crypto@master
	GOFLAGS="" go get gitlab.com/elixxir/client@master

master: update_master clean build version

release: update_release clean build
