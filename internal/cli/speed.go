package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

var speedHistory bool

var speedCmd = &cobra.Command{
	Use:   "speed",
	Short: "Run a speed test",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewSpeedService(runner)

		ctx := context.Background()

		if speedHistory {
			return showSpeedHistory()
		}

		if !jsonFlag {
			fmt.Println()
			fmt.Println("Speed Test")
			fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
			fmt.Println()
			fmt.Println("  Testing download speed (Cloudflare)...")
		}

		result, err := svc.RunSpeedTest(ctx)
		if err != nil {
			return fmt.Errorf("failed to run speed test: %w", err)
		}

		// Populate millisecond fields for JSON.
		result.LatencyMs = float64(result.Latency) / float64(time.Millisecond)

		// Save to history.
		if err := appendSpeedHistory(result); err != nil {
			// Non-fatal: warn but continue.
			if !jsonFlag {
				fmt.Fprintf(os.Stderr, "  Warning: failed to save history: %v\n", err)
			}
		}

		if jsonFlag {
			return printJSON(result)
		}

		fmt.Println()
		fmt.Printf("  Download:  %s\n", network.FormatSpeed(result.DownloadMbps))
		fmt.Printf("  Latency:   %.0fms\n", float64(result.Latency)/float64(time.Millisecond))
		fmt.Printf("  Server:    %s\n", result.Server)
		fmt.Printf("  Time:      %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Println()

		return nil
	},
}

func init() {
	speedCmd.Flags().BoolVar(&speedHistory, "history", false, "Show speed test history")
}

func speedHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "netwhiz", "speed-history.json")
}

func loadSpeedHistory() ([]network.SpeedHistoryEntry, error) {
	path := speedHistoryPath()
	if path == "" {
		return nil, fmt.Errorf("failed to determine home directory")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read speed history: %w", err)
	}

	var history []network.SpeedHistoryEntry
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse speed history: %w", err)
	}

	return history, nil
}

func saveSpeedHistory(history []network.SpeedHistoryEntry) error {
	path := speedHistoryPath()
	if path == "" {
		return fmt.Errorf("failed to determine home directory")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal speed history: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write speed history: %w", err)
	}

	return nil
}

func appendSpeedHistory(result *network.SpeedResult) error {
	history, err := loadSpeedHistory()
	if err != nil {
		// Start fresh if history is corrupt.
		history = nil
	}

	entry := network.SpeedHistoryEntry{
		DownloadMbps: result.DownloadMbps,
		UploadMbps:   result.UploadMbps,
		LatencyMs:    float64(result.Latency) / float64(time.Millisecond),
		Server:       result.Server,
		Timestamp:    result.Timestamp,
	}

	history = append(history, entry)

	// Keep last 100 entries.
	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	return saveSpeedHistory(history)
}

func showSpeedHistory() error {
	history, err := loadSpeedHistory()
	if err != nil {
		return fmt.Errorf("failed to load speed history: %w", err)
	}

	if len(history) == 0 {
		if jsonFlag {
			return printJSON([]network.SpeedHistoryEntry{})
		}
		fmt.Println()
		fmt.Println("  No speed test history found.")
		fmt.Println("  Run 'netwhiz speed' to record a test.")
		fmt.Println()
		return nil
	}

	if jsonFlag {
		return printJSON(history)
	}

	fmt.Println()
	fmt.Println("Speed Test History")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
	fmt.Println()
	fmt.Printf("  %-20s  %12s  %12s  %10s  %s\n", "TIME", "DOWNLOAD", "UPLOAD", "LATENCY", "SERVER")
	fmt.Printf("  %-20s  %12s  %12s  %10s  %s\n", "----", "--------", "------", "-------", "------")

	// Show last 20 entries.
	start := 0
	if len(history) > 20 {
		start = len(history) - 20
	}

	for _, entry := range history[start:] {
		upload := "-"
		if entry.UploadMbps > 0 {
			upload = network.FormatSpeed(entry.UploadMbps)
		}
		fmt.Printf("  %-20s  %12s  %12s  %8.0fms  %s\n",
			entry.Timestamp.Format("2006-01-02 15:04"),
			network.FormatSpeed(entry.DownloadMbps),
			upload,
			entry.LatencyMs,
			entry.Server)
	}

	fmt.Printf("\n  Total entries: %d (showing last %d)\n\n", len(history), len(history)-start)
	return nil
}
