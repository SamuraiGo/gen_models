#!/usr/bin/env bash

echo 'darwin'
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o gen_models_darwin

echo 'linux'
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o gen_models_linux

echo 'windows'
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o gen_models.exe

exit