


fmt:
	go fmt . && go fmt internal/api/* && go fmt internal/app/*.go && go fmt internal/model/*

tidy:
	go mod tidy

run:
	go run cmd/app/main.go

build:
	go build -o http-here cmd/app/main.go

