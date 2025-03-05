


fmt:
	go fmt internal/api/* && go fmt internal/app/*.go && go fmt internal/cert/* && go fmt internal/model/* && go fmt internal/util/*

tidy:
	go mod tidy

run:
	go run cmd/http-here/main.go

build:
	go build -o http-here -ldflags "-w -s" cmd/http-here/main.go


