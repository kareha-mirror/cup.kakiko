all: build

build:
	go build -o skk-edic-merge ./cmd/skk-edic-merge
	make -C skk-edic
	go build -o kakiko ./cmd/kakiko

clean:
	make -C skk-edic clean
	rm -f skk-edic-merge kakiko

run:
	go run ./cmd/kakiko

fmt:
	go fmt ./...

test:
	go test ./...

tidy:
	grep -v '^.tea.kareha.org' go.mod > go.mod.clipped
	mv go.mod.clipped go.mod
	GOPRIVATE=tea.kareha.org go mod tidy

kk:
	mkdir -p config
	./kakiko -joyo -d config
	rm -rf config

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o kakiko.exe ./cmd/kakiko
