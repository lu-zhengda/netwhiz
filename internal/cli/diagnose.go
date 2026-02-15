package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Run a full network diagnostic",
	Long:  "Runs info, dns, wifi, ping, and speed checks and synthesizes a diagnostic report.",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		ctx := context.Background()

		report := &network.DiagnoseReport{
			Timestamp: time.Now(),
		}

		if !jsonFlag {
			fmt.Println()
			fmt.Println("Network Diagnostic Report")
			fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
			fmt.Println()
		}

		// 1. Network Info.
		if !jsonFlag {
			fmt.Println("  [1/5] Checking network info...")
		}
		infoSvc := network.NewInfoService(runner)
		info, err := infoSvc.GetNetworkInfo(ctx)
		if err != nil {
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "network_info",
				Status:  "error",
				Message: fmt.Sprintf("Failed to get network info: %v", err),
			})
		} else {
			status := "ok"
			message := fmt.Sprintf("Interface %s is %s, IP: %s, Gateway: %s", info.Interface, info.Status, info.IPAddress, info.Router)
			if info.Status != "Active" {
				status = "error"
				message = fmt.Sprintf("Interface %s is inactive", info.Interface)
			} else if info.PublicIP == "" {
				status = "warning"
				message += " (public IP unavailable)"
			}
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "network_info",
				Status:  status,
				Message: message,
				Detail:  info,
			})
		}

		// 2. DNS.
		if !jsonFlag {
			fmt.Println("  [2/5] Checking DNS configuration...")
		}
		dnsSvc := network.NewDNSService(runner)
		servers, serviceName, err := dnsSvc.GetDNSServers(ctx)
		if err != nil {
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "dns",
				Status:  "error",
				Message: fmt.Sprintf("Failed to get DNS config: %v", err),
			})
		} else {
			status := "ok"
			message := fmt.Sprintf("DNS service: %s", serviceName)
			if len(servers) > 0 {
				message += fmt.Sprintf(", servers: %v", servers)
			} else {
				status = "warning"
				message += " (using DHCP/auto DNS)"
			}
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "dns",
				Status:  status,
				Message: message,
			})
		}

		// 3. WiFi.
		if !jsonFlag {
			fmt.Println("  [3/5] Checking WiFi...")
		}
		wifiSvc := network.NewWiFiService(runner)
		wifiInfo, err := wifiSvc.GetWiFiInfo(ctx)
		if err != nil {
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "wifi",
				Status:  "warning",
				Message: fmt.Sprintf("WiFi info unavailable: %v", err),
			})
		} else {
			status := "ok"
			quality := network.SignalQuality(wifiInfo.RSSI)
			message := fmt.Sprintf("Connected to %s, RSSI: %ddBm (%s), Channel: %d (%s)",
				wifiInfo.SSID, wifiInfo.RSSI, quality, wifiInfo.Channel, wifiInfo.Band)
			if wifiInfo.RSSI < -75 {
				status = "warning"
				message += " - weak signal"
			}
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "wifi",
				Status:  status,
				Message: message,
				Detail:  wifiInfo,
			})
		}

		// 4. Ping (gateway + 8.8.8.8).
		if !jsonFlag {
			fmt.Println("  [4/5] Testing connectivity (ping)...")
		}
		pingSvc := network.NewPingService(runner)

		// Ping gateway.
		gateway := ""
		if info != nil {
			gateway = info.Router
		}
		if gateway != "" {
			stats, err := pingSvc.Ping(ctx, gateway, 3)
			if err != nil {
				report.Checks = append(report.Checks, network.DiagnoseCheck{
					Name:    "ping_gateway",
					Status:  "error",
					Message: fmt.Sprintf("Failed to ping gateway %s: %v", gateway, err),
				})
			} else {
				status := "ok"
				message := fmt.Sprintf("Gateway %s: avg %.1fms, loss %.0f%%",
					gateway,
					float64(stats.AvgRTT)/float64(time.Millisecond),
					stats.LossPercent)
				if stats.LossPercent > 0 {
					status = "warning"
				}
				if stats.LossPercent >= 100 {
					status = "error"
				}
				report.Checks = append(report.Checks, network.DiagnoseCheck{
					Name:    "ping_gateway",
					Status:  status,
					Message: message,
				})
			}
		}

		// Ping 8.8.8.8.
		stats, err := pingSvc.Ping(ctx, "8.8.8.8", 3)
		if err != nil {
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "ping_internet",
				Status:  "error",
				Message: fmt.Sprintf("Failed to ping 8.8.8.8: %v", err),
			})
		} else {
			status := "ok"
			message := fmt.Sprintf("Internet (8.8.8.8): avg %.1fms, loss %.0f%%",
				float64(stats.AvgRTT)/float64(time.Millisecond),
				stats.LossPercent)
			if stats.LossPercent > 0 {
				status = "warning"
			}
			if stats.LossPercent >= 100 {
				status = "error"
			}
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "ping_internet",
				Status:  status,
				Message: message,
			})
		}

		// 5. Speed test (quick).
		if !jsonFlag {
			fmt.Println("  [5/5] Running speed test...")
		}
		speedSvc := network.NewSpeedService(runner)
		speedResult, err := speedSvc.RunSpeedTest(ctx)
		if err != nil {
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "speed",
				Status:  "warning",
				Message: fmt.Sprintf("Speed test failed: %v", err),
			})
		} else {
			status := "ok"
			message := fmt.Sprintf("Download: %s, Latency: %.0fms",
				network.FormatSpeed(speedResult.DownloadMbps),
				float64(speedResult.Latency)/float64(time.Millisecond))
			if speedResult.DownloadMbps < 10 {
				status = "warning"
				message += " - slow connection"
			}
			report.Checks = append(report.Checks, network.DiagnoseCheck{
				Name:    "speed",
				Status:  status,
				Message: message,
			})
		}

		// Synthesize summary.
		report.Summary = synthesizeSummary(report.Checks)

		if jsonFlag {
			return printJSON(report)
		}

		// Print report.
		fmt.Println()
		for _, check := range report.Checks {
			icon := statusIcon(check.Status)
			fmt.Printf("  %s  %-16s  %s\n", icon, check.Name, check.Message)
		}

		fmt.Println()
		fmt.Printf("  Summary: %s\n", report.Summary)
		fmt.Println()

		return nil
	},
}

func init() {
	// Register in root.go init().
}

func statusIcon(status string) string {
	switch status {
	case "ok":
		return "[OK]"
	case "warning":
		return "[!!]"
	case "error":
		return "[XX]"
	default:
		return "[??]"
	}
}

func synthesizeSummary(checks []network.DiagnoseCheck) string {
	errors := 0
	warnings := 0
	for _, c := range checks {
		switch c.Status {
		case "error":
			errors++
		case "warning":
			warnings++
		}
	}

	if errors > 0 {
		return fmt.Sprintf("Found %d error(s) and %d warning(s). Network connectivity issues detected.", errors, warnings)
	}
	if warnings > 0 {
		return fmt.Sprintf("Found %d warning(s). Network is functional but may have issues.", warnings)
	}
	return "All checks passed. Network is healthy."
}
