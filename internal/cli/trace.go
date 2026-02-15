package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/netwhiz/internal/network"
)

var traceCmd = &cobra.Command{
	Use:   "trace <host>",
	Short: "Traceroute to a host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewTraceService(runner)

		host := args[0]
		ctx := context.Background()

		fmt.Printf("Traceroute to %s\n", host)
		fmt.Println("\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500")
		fmt.Println()

		hops, err := svc.Traceroute(ctx, host)
		if err != nil {
			return fmt.Errorf("failed to traceroute to %s: %w", host, err)
		}

		printTraceroute(hops)
		return nil
	},
}

func printTraceroute(hops []network.TraceHop) {
	for _, hop := range hops {
		if hop.Lost {
			fmt.Printf("  %2d  *  *  *\n", hop.Hop)
			continue
		}

		name := hop.Hostname
		if name == "" {
			name = hop.IP
		}
		if name == "" {
			name = "*"
		}

		var rtts []string
		for _, rtt := range hop.RTTs {
			rtts = append(rtts, fmt.Sprintf("%.1fms", float64(rtt)/float64(time.Millisecond)))
		}

		rttStr := strings.Join(rtts, "  ")
		if hop.IP != "" && hop.IP != hop.Hostname {
			fmt.Printf("  %2d  %s (%s)  %s\n", hop.Hop, name, hop.IP, rttStr)
		} else {
			fmt.Printf("  %2d  %s  %s\n", hop.Hop, name, rttStr)
		}
	}
	fmt.Println()
}
