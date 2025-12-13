package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const protocolICMP = 1 // ICMP protocol number
const serverPort = ":8081"

type Config struct {
	ResponseDelay int // Milliseconds to delay by
}

func main() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))

	//prevent the kernel from responding to pings
	cmd := exec.Command("/bin/sh", "-c", "sudo sysctl -w net.ipv4.icmp_echo_ignore_all=1")
	cmd.Run()

	config := Config{
		ResponseDelay: 20,
	}

	go listenForPings(&config)

	listenForConfigChanges(&config)
}

func listenForConfigChanges(c *Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("Error reading HTTP request body", "error", err)
		}

		slog.Info("Received request", "body", body)

		type Data struct {
			ResponseDelay int
		}
		data := Data{}
		json.Unmarshal(body, &data)

		c.ResponseDelay = data.ResponseDelay
	})

	slog.Info(fmt.Sprintf("Listening on %s", serverPort))

	err := http.ListenAndServe(serverPort, nil)
	if err != nil {
		slog.Error("Error starting HTTP server", "error", err)
	}
}

func listenForPings(c *Config) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		slog.Error("Failed to listen to icmp", "error", err)
	}
	defer conn.Close()

	slog.Info("Listening for icmp requests")

	readBuf := make([]byte, 1500)

	for {
		err := conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			slog.Error("Error setting Read Deadline", "error", err)
		}

		n, addr, err := conn.ReadFrom(readBuf)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				// Timeout is expected, just continue
				slog.Debug("Timeout exceeded")
				continue
			}

			slog.Error("Error reading to buffer", "error", err)
			continue
		}

		rm, err := icmp.ParseMessage(protocolICMP, readBuf[:n])
		if err != nil {
			slog.Error("Error reading icmp message", "error", err)
			continue
		}

		if rm.Type == ipv4.ICMPTypeEcho {
			echo, ok := rm.Body.(*icmp.Echo)
			if !ok {
				slog.Error("Received non-echo body")
			}

			slog.Info(fmt.Sprintf("Received Echo Request from %s (ID: %d, Seq: %d)",
				addr.String(), echo.ID, echo.Seq))

			time.Sleep(time.Duration(c.ResponseDelay) * time.Millisecond)

			wm := icmp.Message{
				Type: ipv4.ICMPTypeEchoReply,
				Code: 0,
				Body: &icmp.Echo{ // Use the same Body (ID, Seq, Data)
					ID:   echo.ID,
					Seq:  echo.Seq,
					Data: echo.Data,
				},
			}

			wb, err := wm.Marshal(nil)
			if err != nil {
				slog.Error("Error marhalling message to bytes")
				continue
			}

			_, err = conn.WriteTo(wb, addr)
			if err != nil {
				slog.Error("Error writing back to addr", err)
				continue
			}

			slog.Info(fmt.Sprintf("Sent Echo Reply to %s", addr.String()))

		} else {
			slog.Error("Not ipv4 echo, received: %s", rm.Type)
		}
	}
}
