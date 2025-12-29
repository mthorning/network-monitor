package monitoring

import (
	"log/slog"
	"slices"
	"sort"
	"strings"
	"sync"
)

type PingTracker struct {
	mu           sync.Mutex
	replies      map[string]bool
	timeouts     []string
	prevTimeouts []string
}

func newPingTracker() *PingTracker {
	return &PingTracker{
		replies:      make(map[string]bool),
		timeouts:     make([]string, 0),
		prevTimeouts: make([]string, 0),
	}
}

func (pt *PingTracker) replyReceived(ip string) {
	pt.mu.Lock()
	pt.replies[ip] = true
	pt.mu.Unlock()
}

func (pt *PingTracker) getTimeouts() []string {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	timeouts := make([]string, 0)
	for ip, replied := range pt.replies {
		if !replied {
			timeouts = append(timeouts, ip)
		}
	}

	sort.Slice(timeouts, func(i, j int) bool {
		return timeouts[i] > timeouts[j]
	})

	if len(pt.prevTimeouts) > 0 && len(timeouts) == 0 {
		slog.Info("No more timeouts")
	} else if len(timeouts) > 0 && !slices.Equal(pt.prevTimeouts, timeouts) {
		slog.Info("Pings timed out", "ips", strings.Join(timeouts, ","))
	}

	pt.prevTimeouts = timeouts
	return timeouts
}

func (pt *PingTracker) reset() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	for ip, _ := range pt.replies {
		pt.replies[ip] = false
	}
	pt.timeouts = make([]string, 0)
	pt.prevTimeouts = make([]string, 0)
}
