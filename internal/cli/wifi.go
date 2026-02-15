package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

var wifiCmd = &cobra.Command{
	Use:   "wifi",
	Short: "Show WiFi information",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewWiFiService(runner)

		ctx := context.Background()
		info, err := svc.GetWiFiInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get WiFi info: %w", err)
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
