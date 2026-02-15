package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show network overview (IP, gateway, DNS, interface)",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewInfoService(runner)

		ctx := context.Background()
		info, err := svc.GetNetworkInfo(ctx)
		if err != nil {
			return fmt.Errorf("failed to get network info: %w", err)
		}

		if jsonFlag {
			return printJSON(info)
		}

		printNetworkInfo(info)
		return nil
	},
}

func printNetworkInfo(info *network.NetworkInfo) {
	fmt.Println()
	fmt.Println("Network Overview")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
	fmt.Println()
	fmt.Printf("  Interface:     %s\n", info.Interface)
	fmt.Printf("  Status:        %s\n", info.Status)
	fmt.Printf("  IP Address:    %s\n", info.IPAddress)
	fmt.Printf("  Subnet Mask:   %s\n", info.SubnetMask)
	fmt.Printf("  Router:        %s\n", info.Router)

	if len(info.DNS) > 0 {
		fmt.Printf("  DNS Servers:   %s\n", strings.Join(info.DNS, ", "))
	} else {
		fmt.Printf("  DNS Servers:   (auto)\n")
	}

	fmt.Printf("  MAC Address:   %s\n", info.MACAddress)

	if info.MediaSpeed != "" {
		fmt.Printf("  Media Speed:   %s\n", info.MediaSpeed)
	}

	fmt.Println()

	if info.PublicIP != "" {
		fmt.Printf("  Public IP:     %s\n", info.PublicIP)
	} else {
		fmt.Printf("  Public IP:     (unavailable)\n")
	}

	fmt.Println()
}
