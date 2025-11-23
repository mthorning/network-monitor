FROM golang:1.25
WORKDIR /app
COPY build build
ENTRYPOINT ./build/network-monitor-dev
