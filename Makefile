all: build

build:
	go build -o kakiko ./cmd/kakiko

clean:
	rm -f kakiko

run:
	go run ./cmd/kakiko

fmt:
	go fmt ./...

test:
	go test ./...
