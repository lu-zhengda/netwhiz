package cli

import (
	"context"
	"fmt"

	"github.com/lu-zhengda/netwhiz/internal/network"
	"github.com/spf13/cobra"
)

var vpnCmd = &cobra.Command{
	Use:   "vpn",
	Short: "VPN management (list, connect, disconnect, status)",
	Long:  "Manage VPN connections. Run without subcommands to list all configured VPNs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewVPNService(runner)

		ctx := context.Background()
		vpns, err := svc.ListVPNs(ctx)
		if err != nil {
			return fmt.Errorf("failed to list VPN connections: %w", err)
		}

		if jsonFlag {
			return printJSON(vpns)
		}

		printVPNList(vpns)
		return nil
	},
}

var vpnConnectCmd = &cobra.Command{
	Use:   "connect <name>",
	Short: "Connect to a VPN by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewVPNService(runner)

		ctx := context.Background()
		name := args[0]

		fmt.Printf("Connecting to VPN %q...\n", name)

		if err := svc.Connect(ctx, name); err != nil {
			return fmt.Errorf("failed to connect to VPN: %w", err)
		}

		// Check status after connecting.
		status, err := svc.GetVPNStatus(ctx, name)
		if err != nil {
			fmt.Printf("Connected (unable to verify status: %v)\n", err)
			return nil
		}

		if status.Status == "Connected" {
			fmt.Printf("Successfully connected to %q\n", name)
		} else {
			fmt.Printf("VPN %q status: %s (may still be connecting)\n", name, status.Status)
		}

		return nil
	},
}

var vpnDisconnectCmd = &cobra.Command{
	Use:   "disconnect [name]",
	Short: "Disconnect VPN (specific or all active)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewVPNService(runner)

		ctx := context.Background()

		if len(args) == 1 {
			name := args[0]
			fmt.Printf("Disconnecting VPN %q...\n", name)

			if err := svc.Disconnect(ctx, name); err != nil {
				return fmt.Errorf("failed to disconnect VPN: %w", err)
			}

			fmt.Printf("Disconnected %q\n", name)
			return nil
		}

		// No name given: disconnect all active VPNs.
		disconnected, err := svc.DisconnectAll(ctx)
		if err != nil {
			return fmt.Errorf("failed to disconnect VPNs: %w", err)
		}

		if len(disconnected) == 0 {
			fmt.Println("No active VPN connections to disconnect.")
		} else {
			for _, name := range disconnected {
				fmt.Printf("Disconnected %q\n", name)
			}
		}

		return nil
	},
}

var vpnStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show detailed status of active VPN",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewVPNService(runner)

		ctx := context.Background()

		var conn *network.VPNConnection
		var err error

		if len(args) == 1 {
			conn, err = svc.GetVPNStatus(ctx, args[0])
		} else {
			conn, err = svc.GetActiveVPNStatus(ctx)
		}

		if err != nil {
			return fmt.Errorf("failed to get VPN status: %w", err)
		}

		if jsonFlag {
			return printJSON(conn)
		}

		printVPNStatus(conn)
		return nil
	},
}

func init() {
	vpnCmd.AddCommand(vpnConnectCmd)
	vpnCmd.AddCommand(vpnDisconnectCmd)
	vpnCmd.AddCommand(vpnStatusCmd)
}

func printVPNList(vpns []network.VPNConnection) {
	fmt.Println()
	fmt.Println("VPN Connections")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
	fmt.Println()

	if len(vpns) == 0 {
		fmt.Println("  No VPN connections configured.")
		fmt.Println()
		return
	}

	fmt.Printf("  %-30s  %-10s  %s\n", "NAME", "TYPE", "STATUS")
	fmt.Printf("  %-30s  %-10s  %s\n", "----", "----", "------")

	for _, vpn := range vpns {
		statusIndicator := "[ ]"
		if vpn.Status == "Connected" {
			statusIndicator = "[*]"
		}
		fmt.Printf("  %-30s  %-10s  %s %s\n",
			truncate(vpn.Name, 30), vpn.Type, statusIndicator, vpn.Status)
	}

	fmt.Println()
}

func printVPNStatus(conn *network.VPNConnection) {
	fmt.Println()
	fmt.Println("VPN Status")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
	fmt.Println()
	fmt.Printf("  Name:        %s\n", conn.Name)
	fmt.Printf("  Type:        %s\n", conn.Type)
	fmt.Printf("  Status:      %s\n", conn.Status)

	if conn.Server != "" {
		fmt.Printf("  Server:      %s\n", conn.Server)
	}
	if conn.RemoteIP != "" {
		fmt.Printf("  Remote IP:   %s\n", conn.RemoteIP)
	}
	if conn.Uptime != "" {
		fmt.Printf("  Uptime:      %s\n", conn.Uptime)
	}

	fmt.Println()
}
