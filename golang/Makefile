.PHONY: default build test clean run server

default: run

run:
	go run tracker.go

build:
	go build -o tracker tracker.go

test:
	go test ./...

clean:
	rm -f tracker

server:
	go build -o tracker tracker.go
	./tracker server

