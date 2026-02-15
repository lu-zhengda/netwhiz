package network

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNetworkInfoJSON(t *testing.T) {
	info := NetworkInfo{
		Interface:  "en0",
		IPAddress:  "192.168.1.100",
		SubnetMask: "255.255.255.0",
		Router:     "192.168.1.1",
		DNS:        []string{"1.1.1.1", "8.8.8.8"},
		MACAddress: "aa:bb:cc:dd:ee:ff",
		Status:     "Active",
		PublicIP:   "1.2.3.4",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"interface", "ip_address", "subnet_mask", "router", "dns", "mac_address", "status", "public_ip"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestWiFiInfoJSON(t *testing.T) {
	info := WiFiInfo{
		SSID:        "MyNetwork",
		BSSID:       "aa:bb:cc:dd:ee:ff",
		Channel:     36,
		Band:        "5GHz",
		RSSI:        -55,
		Noise:       -90,
		SNR:         35,
		TxRate:      400,
		Security:    "WPA3",
		CountryCode: "US",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"ssid", "bssid", "channel", "band", "rssi", "noise", "snr", "tx_rate", "security", "country_code"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestPingStatsJSON(t *testing.T) {
	stats := PingStats{
		Host:        "google.com",
		Sent:        5,
		Received:    5,
		Lost:        0,
		LossPercent: 0,
		MinRTT:      10 * time.Millisecond,
		MinRTTMs:    10.0,
		MaxRTT:      50 * time.Millisecond,
		MaxRTTMs:    50.0,
		AvgRTT:      25 * time.Millisecond,
		AvgRTTMs:    25.0,
		StdDevRTT:   5 * time.Millisecond,
		StdDevRTTMs: 5.0,
		Results: []PingResult{
			{Host: "google.com", Seq: 0, TTL: 117, Time: 10 * time.Millisecond, TimeMs: 10.0},
		},
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"host", "sent", "received", "lost", "loss_percent", "min_rtt_ms", "max_rtt_ms", "avg_rtt_ms", "results"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestTraceHopJSON(t *testing.T) {
	hop := TraceHop{
		Hop:      1,
		IP:       "192.168.1.1",
		Hostname: "router.local",
		RTTs:     []time.Duration{10 * time.Millisecond, 11 * time.Millisecond},
		RTTsMs:   []float64{10.0, 11.0},
		Lost:     false,
	}

	data, err := json.Marshal(hop)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"hop", "ip", "hostname", "rtts_ms", "lost"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestSpeedResultJSON(t *testing.T) {
	result := SpeedResult{
		DownloadMbps: 150.5,
		UploadMbps:   50.2,
		Latency:      12 * time.Millisecond,
		LatencyMs:    12.0,
		Server:       "Cloudflare",
		Timestamp:    time.Now(),
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"download_mbps", "upload_mbps", "latency_ms", "server", "timestamp"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestARPEntryJSON(t *testing.T) {
	entry := ARPEntry{
		IP:        "192.168.1.100",
		MAC:       "aa:bb:cc:dd:ee:ff",
		Interface: "en0",
		Hostname:  "myhost.local",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"ip", "mac", "interface", "hostname"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestDNSBenchmarkResultJSON(t *testing.T) {
	result := DNSBenchmarkResult{
		Provider:  "Cloudflare",
		Server:    "1.1.1.1",
		LatencyMs: 5.5,
		Reachable: true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"provider", "server", "latency_ms", "reachable"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestDiagnoseCheckJSON(t *testing.T) {
	check := DiagnoseCheck{
		Name:    "network_info",
		Status:  "ok",
		Message: "all good",
		Detail:  map[string]string{"foo": "bar"},
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"name", "status", "message", "detail"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestWiFiSignalEventJSON(t *testing.T) {
	event := WiFiSignalEvent{
		SSID:      "MyNetwork",
		RSSI:      -55,
		Noise:     -90,
		SNR:       35,
		Quality:   "Good",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"ssid", "rssi", "noise", "snr", "quality", "timestamp"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestDNSPresetJSON(t *testing.T) {
	preset := DNSPreset{
		Name:    "cloudflare",
		Servers: []string{"1.1.1.1", "1.0.0.1"},
	}

	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["name"] != "cloudflare" {
		t.Errorf("expected name=cloudflare, got %v", parsed["name"])
	}
}
