package network

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// SpeedService measures network speed.
type SpeedService struct {
	runner CmdRunner
}

// NewSpeedService creates a new SpeedService.
func NewSpeedService(runner CmdRunner) *SpeedService {
	return &SpeedService{runner: runner}
}

// OoklaAvailable checks if the Ookla speedtest CLI is installed.
func OoklaAvailable() bool {
	_, err := exec.LookPath("speedtest")
	return err == nil
}

// RunSpeedTest performs a download speed test using the default endpoint.
func (s *SpeedService) RunSpeedTest(ctx context.Context) (*SpeedResult, error) {
	return s.RunSpeedTestWith(ctx, SpeedEndpoints[0])
}

// RunSpeedTestWith performs a speed test using a specific endpoint.
// If the endpoint is the Ookla endpoint and the CLI is installed, it uses
// the Ookla protocol. Otherwise it falls back to curl-based testing.
func (s *SpeedService) RunSpeedTestWith(ctx context.Context, endpoint SpeedEndpoint) (*SpeedResult, error) {
	if endpoint.Ookla {
		return s.runOoklaTest(ctx)
	}
	return s.runCurlTest(ctx, endpoint)
}

// runOoklaTest uses the official Ookla speedtest CLI for accurate gigabit testing.
func (s *SpeedService) runOoklaTest(ctx context.Context) (*SpeedResult, error) {
	out, err := s.runner.Run(ctx, "speedtest", "--format=json", "--accept-license", "--accept-gdpr")
	if err != nil {
		return nil, fmt.Errorf("speedtest CLI failed: %w", err)
	}

	var raw ooklaResult
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse speedtest output: %w", err)
	}

	serverName := raw.Server.Name
	if raw.Server.Location != "" {
		serverName += " (" + raw.Server.Location + ")"
	}

	return &SpeedResult{
		DownloadMbps: float64(raw.Download.Bandwidth*8) / 1_000_000,
		UploadMbps:   float64(raw.Upload.Bandwidth*8) / 1_000_000,
		Latency:      time.Duration(raw.Ping.Latency * float64(time.Millisecond)),
		Server:       serverName,
		Timestamp:    time.Now(),
	}, nil
}

// ooklaResult is the JSON structure from `speedtest --format=json`.
type ooklaResult struct {
	Download struct {
		Bandwidth int64 `json:"bandwidth"` // bytes/sec
	} `json:"download"`
	Upload struct {
		Bandwidth int64 `json:"bandwidth"` // bytes/sec
	} `json:"upload"`
	Ping struct {
		Latency float64 `json:"latency"` // milliseconds
	} `json:"ping"`
	Server struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	} `json:"server"`
}

// runCurlTest performs a curl-based speed test (Cloudflare endpoints).
func (s *SpeedService) runCurlTest(ctx context.Context, endpoint SpeedEndpoint) (*SpeedResult, error) {
	result := &SpeedResult{
		Server:    endpoint.Name,
		Timestamp: time.Now(),
	}

	// Measure latency first with a small request.
	latencyStart := time.Now()
	_, err := s.runner.Run(ctx, "curl", "-s", "--max-time", "5", "-o", "/dev/null", "https://speed.cloudflare.com/__down?bytes=1000")
	if err != nil {
		return nil, fmt.Errorf("failed to measure latency: %w", err)
	}
	result.Latency = time.Since(latencyStart)

	// Measure download speed.
	// curl -o /dev/null -w "%{speed_download}" outputs bytes per second.
	out, err := s.runner.Run(ctx, "curl", "-s", "--max-time", "30", "-o", "/dev/null", "-w", "%{speed_download}", endpoint.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to run speed test: %w", err)
	}

	bytesPerSec, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse speed result: %w", err)
	}

	// Convert bytes/sec to Mbps (megabits per second).
	result.DownloadMbps = (bytesPerSec * 8) / 1_000_000

	// Measure upload speed if the endpoint supports it.
	if endpoint.UploadURL != "" {
		uploadCmd := fmt.Sprintf(
			`dd if=/dev/urandom bs=1000000 count=5 2>/dev/null | curl -s --max-time 15 -X POST -o /dev/null -w "%%{speed_upload}" --data-binary @- %s`,
			endpoint.UploadURL)
		upOut, err := s.runner.Run(ctx, "bash", "-c", uploadCmd)
		if err == nil {
			upBytesPerSec, err := strconv.ParseFloat(strings.TrimSpace(string(upOut)), 64)
			if err == nil {
				result.UploadMbps = (upBytesPerSec * 8) / 1_000_000
			}
		}
	}

	return result, nil
}

// FormatSpeed returns a human-readable speed string.
func FormatSpeed(mbps float64) string {
	if mbps >= 1000 {
		return fmt.Sprintf("%.1f Gbps", mbps/1000)
	}
	return fmt.Sprintf("%.1f Mbps", mbps)
}
