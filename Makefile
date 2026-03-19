.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o build/network-monitor ./cmd/network_monitor/main.go

.PHONY: build-amd
build-amd:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o build/network-monitor ./cmd/network_monitor/main.go

.PHONY: dev
dev:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/icmp_responder ./cmd/icmp_responder/main.go

.PHONY: test
test:
	go test -v ./internal/utils

.PHONY: ping-vm
ping-vm:
	go run cmd/network-monitor/main.go -ping-ips "192.168.139.205" -log-level "DEBUG"

.PHONY: clean
clean:
	rm -rf ./build

.PHONY: release
release:
	$(MAKE) clean
	$(MAKE) build-amd
	scp -p ./build/network-monitor monitoring:/tmp/network-monitor
	ssh -t monitoring "bash -c 'mv /tmp/network-monitor /usr/local/bin && chmod 0755 /usr/local/bin/network-monitor && systemctl restart network-monitor'"
