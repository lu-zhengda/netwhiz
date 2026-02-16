package cli

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
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

		if jsonFlag {
			config := network.DNSConfig{
				ServiceName: serviceName,
				Servers:     servers,
				Presets:     network.DNSPresets,
			}
			for _, p := range network.DNSPresets {
				if matchesDNSPreset(servers, p.Servers) {
					config.Preset = p.Name
					break
				}
			}
			return printJSON(config)
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
			if matchesDNSPreset(servers, p.Servers) {
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

		if !jsonFlag {
			fmt.Printf("Setting DNS to %s...\n", server)
		}

		if err := svc.SetDNS(ctx, server); err != nil {
			return fmt.Errorf("failed to set DNS: %w", err)
		}

		if jsonFlag {
			msg := "DNS set to " + server
			if preset := network.ResolveDNSPreset(server); preset != nil {
				msg = fmt.Sprintf("DNS set to %s (%s)", server, strings.Join(preset, ", "))
			}
			return printJSON(jsonAction{
				OK:      true,
				Action:  "dns_set",
				Target:  server,
				Message: msg,
			})
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

		if !jsonFlag {
			fmt.Println("Flushing DNS cache...")
		}

		if err := svc.FlushDNS(ctx); err != nil {
			return fmt.Errorf("failed to flush DNS cache: %w", err)
		}

		if jsonFlag {
			return printJSON(jsonAction{
				OK:      true,
				Action:  "dns_flush",
				Message: "DNS cache flushed successfully",
			})
		}

		fmt.Println("DNS cache flushed successfully.")
		fmt.Println("\nNote: For full DNS flush, you may also need:")
		fmt.Println("  sudo killall -HUP mDNSResponder")

		return nil
	},
}

var dnsBenchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark DNS providers by response time",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewDNSService(runner)

		ctx := context.Background()

		// Get current DNS for comparison.
		currentServers, _, _ := svc.GetDNSServers(ctx)

		report := runDNSBenchmark(currentServers)

		if jsonFlag {
			return printJSON(report)
		}

		printDNSBenchmark(report)
		return nil
	},
}

func init() {
	dnsCmd.AddCommand(dnsSetCmd)
	dnsCmd.AddCommand(dnsFlushCmd)
	dnsCmd.AddCommand(dnsBenchmarkCmd)
}

// matchesDNSPreset checks if the given servers match a DNS preset.
func matchesDNSPreset(servers, preset []string) bool {
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

// runDNSBenchmark tests DNS providers and returns a benchmark report.
func runDNSBenchmark(currentServers []string) *network.DNSBenchmarkReport {
	type benchTarget struct {
		name   string
		server string
	}

	targets := []benchTarget{
		{name: "Cloudflare", server: "1.1.1.1"},
		{name: "Google", server: "8.8.8.8"},
		{name: "Quad9", server: "9.9.9.9"},
	}

	// Add current DNS if it's not already one of the above.
	for _, s := range currentServers {
		found := false
		for _, t := range targets {
			if t.server == s {
				found = true
				break
			}
		}
		if !found && s != "" {
			targets = append(targets, benchTarget{name: "Current (" + s + ")", server: s})
		}
	}

	var results []network.DNSBenchmarkResult

	for _, t := range targets {
		result := benchmarkDNS(t.name, t.server)
		results = append(results, result)
	}

	// Sort by latency (unreachable last).
	sort.Slice(results, func(i, j int) bool {
		if !results[i].Reachable && !results[j].Reachable {
			return false
		}
		if !results[i].Reachable {
			return false
		}
		if !results[j].Reachable {
			return true
		}
		return results[i].LatencyMs < results[j].LatencyMs
	})

	report := &network.DNSBenchmarkReport{
		Results:  results,
		TestedAt: time.Now(),
	}

	if len(results) > 0 && results[0].Reachable {
		report.Fastest = results[0].Provider
		report.FastestMs = results[0].LatencyMs
	}

	return report
}

// benchmarkDNS measures DNS lookup time for a single server.
func benchmarkDNS(name, server string) network.DNSBenchmarkResult {
	result := network.DNSBenchmarkResult{
		Provider: name,
		Server:   server,
	}

	// Use net.Resolver with custom dialer to test specific DNS server.
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, netw, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: 3 * time.Second}
			return d.DialContext(ctx, "udp", server+":53")
		},
	}

	// Run 3 lookups and take the average.
	const numTests = 3
	var totalMs float64
	testDomain := "google.com"

	for i := 0; i < numTests; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		start := time.Now()
		_, err := resolver.LookupHost(ctx, testDomain)
		elapsed := time.Since(start)
		cancel()

		if err != nil {
			result.Error = fmt.Sprintf("failed to resolve %s: %v", testDomain, err)
			result.Reachable = false
			return result
		}

		totalMs += float64(elapsed) / float64(time.Millisecond)
	}

	result.LatencyMs = totalMs / float64(numTests)
	result.Reachable = true
	return result
}

func printDNSBenchmark(report *network.DNSBenchmarkReport) {
	fmt.Println()
	fmt.Println("DNS Benchmark")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
	fmt.Println()
	fmt.Printf("  %-25s  %-15s  %10s  %s\n", "PROVIDER", "SERVER", "LATENCY", "STATUS")
	fmt.Printf("  %-25s  %-15s  %10s  %s\n", "--------", "------", "-------", "------")

	for _, r := range report.Results {
		status := "ok"
		latency := fmt.Sprintf("%.1fms", r.LatencyMs)
		if !r.Reachable {
			status = "error"
			latency = "-"
		}
		fmt.Printf("  %-25s  %-15s  %10s  %s\n", r.Provider, r.Server, latency, status)
	}

	fmt.Println()
	if report.Fastest != "" {
		fmt.Printf("  Recommended: %s (%.1fms avg)\n", report.Fastest, report.FastestMs)
	} else {
		fmt.Println("  No reachable DNS servers found.")
	}
	fmt.Println()
}
