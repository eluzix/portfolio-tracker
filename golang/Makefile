.PHONY: default build test clean run

default: run

run:
	go run tracker.go

build:
	go build -o tracker tracker.go

test:
	go test ./...

clean:
	rm -f tracker
