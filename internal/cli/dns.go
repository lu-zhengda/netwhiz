package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/netwhiz/internal/network"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Show current DNS servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewDNSService(runner)

		ctx := context.Background()
		servers, serviceName, err := svc.GetDNSServers(ctx)
		if err != nil {
			return fmt.Errorf("failed to get DNS servers: %w", err)
		}

		fmt.Println()
		fmt.Println("DNS Configuration")
		fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
		fmt.Println()
		fmt.Printf("  Network Service:  %s\n", serviceName)

		if len(servers) > 0 {
			fmt.Printf("  DNS Servers:      %s\n", strings.Join(servers, ", "))
		} else {
			fmt.Printf("  DNS Servers:      (auto / DHCP)\n")
		}

		// Show known presets.
		for _, p := range network.DNSPresets {
			if matchesPreset(servers, p.Servers) {
				fmt.Printf("  Preset:           %s\n", p.Name)
				break
			}
		}

		fmt.Println()
		fmt.Println("  Available presets:")
		for _, p := range network.DNSPresets {
			fmt.Printf("    %-12s %s\n", p.Name, strings.Join(p.Servers, ", "))
		}
		fmt.Println()

		return nil
	},
}

var dnsSetCmd = &cobra.Command{
	Use:   "set <server>",
	Short: "Set DNS server (e.g., 'cloudflare', 'google', '1.1.1.1')",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewDNSService(runner)

		ctx := context.Background()
		server := args[0]

		fmt.Printf("Setting DNS to %s...\n", server)

		if err := svc.SetDNS(ctx, server); err != nil {
			return fmt.Errorf("failed to set DNS: %w", err)
		}

		// Show what was set.
		if preset := network.ResolveDNSPreset(server); preset != nil {
			fmt.Printf("DNS set to %s (%s)\n", server, strings.Join(preset, ", "))
		} else {
			fmt.Printf("DNS set to %s\n", server)
		}

		fmt.Println("\nNote: If this failed, try running with sudo:")
		fmt.Println("  sudo netwhiz dns set", server)

		return nil
	},
}

var dnsFlushCmd = &cobra.Command{
	Use:   "flush",
	Short: "Flush DNS cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewDNSService(runner)

		ctx := context.Background()
		fmt.Println("Flushing DNS cache...")

		if err := svc.FlushDNS(ctx); err != nil {
			return fmt.Errorf("failed to flush DNS cache: %w", err)
		}

		fmt.Println("DNS cache flushed successfully.")
		fmt.Println("\nNote: For full DNS flush, you may also need:")
		fmt.Println("  sudo killall -HUP mDNSResponder")

		return nil
	},
}

func init() {
	dnsCmd.AddCommand(dnsSetCmd)
	dnsCmd.AddCommand(dnsFlushCmd)
}

// matchesPreset checks if the given servers match a DNS preset.
func matchesPreset(servers, preset []string) bool {
	if len(servers) != len(preset) {
		return false
	}
	for i := range servers {
		if servers[i] != preset[i] {
			return false
		}
	}
	return true
}
