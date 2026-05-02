all: build

build:
	go build -o kakiko ./cmd/kakiko
	go build -o joyo ./cmd/joyo

clean:
	rm -f kakiko joyo

run:
	go run ./cmd/kakiko

fmt:
	go fmt ./...

test:
	go test ./...
