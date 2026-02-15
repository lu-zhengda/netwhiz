package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lu-zhengda/netwhiz/internal/network"
)

func TestSpeedHistoryPath(t *testing.T) {
	path := speedHistoryPath()
	if path == "" {
		t.Skip("unable to determine home directory")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %s", path)
	}
	if filepath.Base(path) != "speed-history.json" {
		t.Errorf("expected speed-history.json, got %s", filepath.Base(path))
	}
}

func TestSpeedHistorySaveLoad(t *testing.T) {
	// Create a temp dir for history file.
	tmpDir := t.TempDir()
	historyPath := filepath.Join(tmpDir, "speed-history.json")

	entries := []network.SpeedHistoryEntry{
		{
			DownloadMbps: 100.5,
			UploadMbps:   50.2,
			LatencyMs:    12.3,
			Server:       "Test Server",
			Timestamp:    time.Now(),
		},
		{
			DownloadMbps: 200.0,
			UploadMbps:   80.0,
			LatencyMs:    8.5,
			Server:       "Test Server 2",
			Timestamp:    time.Now(),
		},
	}

	// Save.
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := os.WriteFile(historyPath, data, 0o644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Load.
	readData, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	var loaded []network.SpeedHistoryEntry
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded))
	}

	if loaded[0].DownloadMbps != 100.5 {
		t.Errorf("expected download 100.5, got %f", loaded[0].DownloadMbps)
	}

	if loaded[1].Server != "Test Server 2" {
		t.Errorf("expected server 'Test Server 2', got %s", loaded[1].Server)
	}
}

func TestSpeedHistoryEntryJSON(t *testing.T) {
	entry := network.SpeedHistoryEntry{
		DownloadMbps: 150.5,
		UploadMbps:   75.3,
		LatencyMs:    10.2,
		Server:       "Cloudflare",
		Timestamp:    time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(entry)
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
			t.Errorf("expected key %s in JSON output", key)
		}
	}
}
