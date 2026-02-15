package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/netwhiz/internal/network"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "ARP scan local network",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewScanService(runner)

		ctx := context.Background()

		fmt.Println()
		fmt.Println("ARP Network Scan")
		fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
		fmt.Println()

		entries, err := svc.ARPScan(ctx)
		if err != nil {
			return fmt.Errorf("failed to scan network: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("  No devices found.")
			return nil
		}

		fmt.Printf("  %-16s  %-19s  %-6s  %s\n", "IP", "MAC", "IFACE", "HOSTNAME")
		fmt.Printf("  %-16s  %-19s  %-6s  %s\n", "----", "---", "-----", "--------")

		for _, e := range entries {
			hostname := e.Hostname
			if hostname == "" {
				hostname = "-"
			}
			fmt.Printf("  %-16s  %-19s  %-6s  %s\n", e.IP, e.MAC, e.Interface, hostname)
		}

		fmt.Printf("\n  Found %d devices\n\n", len(entries))
		return nil
	},
}
