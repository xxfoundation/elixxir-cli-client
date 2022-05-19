#!/bin/bash

GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o cli-client.linux64 main.go
GOOS=windows GOARCH=amd64 go build -ldflags '-w -s' -o cli-client.win64 main.go
GOOS=darwin GOARCH=amd64 go build -ldflags '-w -s' -o cli-client.darwin64 main.go