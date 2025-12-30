package monitoring

import (
	"network_monitor/internal/utils"
)

type timeoutTracker struct {
	replies      *utils.Tracker[bool]
	timeoutCount *utils.Tracker[int]
}

func newTimeoutTracker(ips []string) *timeoutTracker {
	tt := timeoutTracker{
		replies:      utils.NewTracker[bool](),
		timeoutCount: utils.NewTracker[int](),
	}

	for _, ip := range ips {
		tt.replies.Set(ip, false)
		tt.timeoutCount.Set(ip, 0)
	}

	return &tt
}

func (tt *timeoutTracker) replyReceived(ip string) {
	tt.replies.Set(ip, true)
}

type timeout struct {
	ip    string
	count int
}

func (tt *timeoutTracker) getTimeouts() []timeout {
	timeouts := make([]timeout, 0)
	replies := tt.replies.GetAll()
	for ip, replied := range replies {
		if !replied {
			count := tt.timeoutCount.Get(ip) + 1
			tt.timeoutCount.Set(ip, count)

			t := timeout{ip, count}
			timeouts = append(timeouts, t)
		} else {
			tt.timeoutCount.Set(ip, 0)
		}
	}

	tt.replies.SetAll(false)

	return timeouts
}

func (tt *timeoutTracker) resetCount(ip string) {
	tt.timeoutCount.Set(ip, 0)
}
