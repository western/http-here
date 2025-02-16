


fmt:
	go fmt . && go fmt controller/* && go fmt model/* && go fmt conf/*

tidy:
	go mod tidy

run:
	go run cmd/app/main.go

build:
	go build -o http-here cmd/app/main.go

