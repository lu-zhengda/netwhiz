package cli

import (
	"testing"
	"time"

	"github.com/lu-zhengda/netwhiz/internal/network"
)

func TestPopulatePingMs(t *testing.T) {
	stats := &network.PingStats{
		Host:      "google.com",
		Sent:      3,
		Received:  3,
		Lost:      0,
		MinRTT:    10 * time.Millisecond,
		MaxRTT:    50 * time.Millisecond,
		AvgRTT:    25 * time.Millisecond,
		StdDevRTT: 5 * time.Millisecond,
		Results: []network.PingResult{
			{Host: "google.com", Seq: 0, TTL: 117, Time: 10 * time.Millisecond},
			{Host: "google.com", Seq: 1, TTL: 117, Time: 25 * time.Millisecond},
			{Host: "google.com", Seq: 2, TTL: 117, Time: 50 * time.Millisecond},
		},
	}

	populatePingMs(stats)

	if stats.MinRTTMs != 10.0 {
		t.Errorf("expected MinRTTMs=10.0, got %f", stats.MinRTTMs)
	}
	if stats.MaxRTTMs != 50.0 {
		t.Errorf("expected MaxRTTMs=50.0, got %f", stats.MaxRTTMs)
	}
	if stats.AvgRTTMs != 25.0 {
		t.Errorf("expected AvgRTTMs=25.0, got %f", stats.AvgRTTMs)
	}
	if stats.StdDevRTTMs != 5.0 {
		t.Errorf("expected StdDevRTTMs=5.0, got %f", stats.StdDevRTTMs)
	}
	if stats.Results[0].TimeMs != 10.0 {
		t.Errorf("expected Results[0].TimeMs=10.0, got %f", stats.Results[0].TimeMs)
	}
	if stats.Results[2].TimeMs != 50.0 {
		t.Errorf("expected Results[2].TimeMs=50.0, got %f", stats.Results[2].TimeMs)
	}
}
