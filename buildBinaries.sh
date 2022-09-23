#!/bin/bash

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '-w -s' -o cli-client.linux64 main.go
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '-w -s' -o cli-client.exe main.go
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags '-w -s' -o cli-client.darwin64 main.go
