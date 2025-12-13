.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/network-monitor ./cmd/network-monitor/main.go

.PHONY: dev
dev:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/icmp-responder ./cmd/icmp-responder/main.go

.PHONY: clean
clean:
	rm -rf ./build
