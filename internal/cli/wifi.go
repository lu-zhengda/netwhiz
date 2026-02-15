package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

var (
	wifiMonitor  bool
	wifiInterval int
	wifiMinRSSI  int
)

var wifiCmd = &cobra.Command{
	Use:   "wifi",
	Short: "Show WiFi information",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewWiFiService(runner)

		ctx := context.Background()

		if wifiMonitor {
			return runWiFiMonitor(ctx, svc)
		}

		info, err := svc.GetWiFiInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get WiFi info: %w", err)
		}

		if jsonFlag {
			return printJSON(info)
		}

		printWiFiInfo(info)
		return nil
	},
}

var wifiScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan nearby WiFi networks",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewWiFiService(runner)

		ctx := context.Background()
		networks, err := svc.ScanNetworks(ctx)
		if err != nil {
			return fmt.Errorf("failed to scan WiFi networks: %w", err)
		}

		if jsonFlag {
			return printJSON(networks)
		}

		fmt.Println()
		fmt.Println("Nearby WiFi Networks")
		fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
		fmt.Println()

		if len(networks) == 0 {
			fmt.Println("  No networks found.")
			return nil
		}

		fmt.Printf("  %-30s  %6s  %4s  %s\n", "SSID", "RSSI", "CH", "SECURITY")
		fmt.Printf("  %-30s  %6s  %4s  %s\n", "----", "----", "--", "--------")

		for _, n := range networks {
			signal := network.SignalBar(n.RSSI)
			fmt.Printf("  %-30s  %4ddBm  %4d  %s  %s\n",
				truncate(n.SSID, 30), n.RSSI, n.Channel, signal, n.Security)
		}

		fmt.Println()
		return nil
	},
}

func init() {
	wifiCmd.Flags().BoolVar(&wifiMonitor, "monitor", false, "Monitor WiFi signal strength continuously")
	wifiCmd.Flags().IntVar(&wifiInterval, "interval", 5, "Polling interval in seconds (used with --monitor)")
	wifiCmd.Flags().IntVar(&wifiMinRSSI, "min-rssi", -75, "Minimum RSSI threshold in dBm (used with --monitor)")
	wifiCmd.AddCommand(wifiScanCmd)
}

func printWiFiInfo(info *network.WiFiInfo) {
	fmt.Println()
	fmt.Println("WiFi Information")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
	fmt.Println()
	fmt.Printf("  SSID:          %s\n", info.SSID)
	fmt.Printf("  BSSID:         %s\n", info.BSSID)
	fmt.Printf("  Channel:       %d (%s)\n", info.Channel, info.Band)
	fmt.Printf("  RSSI:          %d dBm (%s)\n", info.RSSI, network.SignalQuality(info.RSSI))
	fmt.Printf("  Noise:         %d dBm\n", info.Noise)
	fmt.Printf("  SNR:           %d dB\n", info.SNR)
	fmt.Printf("  Tx Rate:       %d Mbps\n", info.TxRate)
	fmt.Printf("  Security:      %s\n", info.Security)
	fmt.Printf("  Country:       %s\n", info.CountryCode)
	fmt.Println()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// runWiFiMonitor polls WiFi signal strength and alerts if below threshold.
func runWiFiMonitor(ctx context.Context, svc *network.WiFiService) error {
	interval := time.Duration(wifiInterval) * time.Second
	if interval < time.Second {
		interval = time.Second
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	belowThreshold := false

	if !jsonFlag {
		fmt.Println()
		fmt.Println("WiFi Signal Monitor")
		fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
		fmt.Printf("  Interval: %ds  Min RSSI: %ddBm\n", wifiInterval, wifiMinRSSI)
		fmt.Println()
		fmt.Printf("  %-20s  %6s  %6s  %4s  %s\n", "TIME", "RSSI", "NOISE", "SNR", "QUALITY")
		fmt.Printf("  %-20s  %6s  %6s  %4s  %s\n", "----", "----", "-----", "---", "-------")
	}

	var events []network.WiFiSignalEvent
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately first, then on each tick.
	poll := func() bool {
		info, err := svc.GetWiFiInfo(ctx)
		if err != nil {
			if !jsonFlag {
				fmt.Printf("  %-20s  Error: %v\n", time.Now().Format("15:04:05"), err)
			}
			return false
		}

		event := network.WiFiSignalEvent{
			SSID:      info.SSID,
			RSSI:      info.RSSI,
			Noise:     info.Noise,
			SNR:       info.SNR,
			Quality:   network.SignalQuality(info.RSSI),
			Timestamp: time.Now(),
		}
		events = append(events, event)

		if jsonFlag {
			// In JSON mode, events are collected and printed at exit.
		} else {
			bar := network.SignalBar(info.RSSI)
			alert := ""
			if info.RSSI < wifiMinRSSI {
				alert = "  ** BELOW THRESHOLD **"
			}
			fmt.Printf("  %-20s  %4ddBm  %4ddBm  %3ddB  %s %s%s\n",
				time.Now().Format("15:04:05"),
				info.RSSI, info.Noise, info.SNR,
				event.Quality, bar, alert)
		}

		if info.RSSI < wifiMinRSSI {
			belowThreshold = true
		}

		return false
	}

	poll()
	for {
		select {
		case <-ticker.C:
			poll()
		case <-sigCh:
			if jsonFlag {
				_ = printJSON(events)
			} else {
				fmt.Println()
				fmt.Printf("  Stopped. Collected %d samples.\n\n", len(events))
			}
			if belowThreshold {
				os.Exit(1)
			}
			return nil
		case <-ctx.Done():
			if jsonFlag {
				_ = printJSON(events)
			}
			if belowThreshold {
				os.Exit(1)
			}
			return nil
		}
	}
}
