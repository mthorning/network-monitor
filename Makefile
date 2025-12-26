.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o build/network-monitor ./cmd/network-monitor/main.go

.PHONY: dev
dev:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/icmp-responder ./cmd/icmp-responder/main.go

.PHONY: clean
clean:
	rm -rf ./build

.PHONY: release
release:
	$(MAKE) clean
	$(MAKE) build
	scp ./build/network-monitor pi:/tmp/network-monitor
	ssh -t pi "sudo bash -c 'mv /tmp/network-monitor /usr/local/bin && sudo systemctl restart network-monitor'"
