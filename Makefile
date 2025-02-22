


fmt:
	go fmt internal/api/* && go fmt internal/app/*.go && go fmt internal/model/*

tidy:
	go mod tidy

run:
	go run cmd/http-here/main.go

build:
	go build -o http-here cmd/http-here/main.go

