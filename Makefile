all: build

build:
	go build -o skk-edic-merge ./cmd/skk-edic-merge
	make -C skk-edic
	go build -o kakiko ./cmd/kakiko
	go build -tags minimal -o kakikom ./cmd/kakiko

clean:
	make -C skk-edic clean
	rm -f kakiko skk-edic-merge

run:
	go run ./cmd/kakiko

fmt:
	go fmt ./...

test:
	go test ./...
