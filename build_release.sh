#!/bin/bash


set -x


rm http-here http-here.exe http-here.darwin.amd64.gz http-here.freebsd.amd64.gz http-here.gz http-here.zip


GOOS=linux GOARCH=amd64 go build -o http-here -ldflags "-w -s" cmd/http-here/main.go
gzip -9 http-here


GOOS=windows GOARCH=amd64 go build -o http-here.exe -ldflags "-w -s" cmd/http-here/main.go
zip -9 http-here.zip http-here.exe
rm http-here.exe


GOOS=darwin GOARCH=amd64 go build -o http-here -ldflags "-w -s" cmd/http-here/main.go
gzip -S .darwin.amd64.gz -9 http-here


GOOS=freebsd GOARCH=amd64 go build -o http-here -ldflags "-w -s" cmd/http-here/main.go
gzip -S .freebsd.amd64.gz -9 http-here

