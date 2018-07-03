@echo off
@color 06

set GoDevWork=%cd%\

echo "Build For windows ..."
set GOOS=windows
set GOARCH=amd64
set GOPATH=%GoDevWork%;%GOPATH%
go build -ldflags "-s -w" -o gen_models.exe
echo "--------- Build For windows Success!"

echo "Build For linux ..."
set GOOS=linux
set GOARCH=amd64
set GOPATH=%GoDevWork%;%GOPATH%
go build -ldflags "-s -w" -o gen_models_linux
echo "--------- Build For linux Success!"

echo "Build For darwin ..."
set GOOS=darwin
set GOARCH=amd64
set GOPATH=%GoDevWork%;%GOPATH%
go build -ldflags "-s -w" -o gen_models_darwin
echo "--------- Build For darwin Success!"

pause