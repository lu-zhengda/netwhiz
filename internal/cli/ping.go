package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

var pingCount int

var pingCmd = &cobra.Command{
	Use:   "ping <host>",
	Short: "Enhanced ping with stats",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewPingService(runner)

		host := args[0]
		ctx := context.Background()

		fmt.Printf("PING %s\n", host)
		fmt.Println("\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500")

		stats, err := svc.Ping(ctx, host, pingCount)
		if err != nil {
			return fmt.Errorf("failed to ping %s: %w", host, err)
		}

		printPingResults(stats)
		return nil
	},
}

func init() {
	pingCmd.Flags().IntVarP(&pingCount, "count", "c", 5, "Number of pings to send")
}

func printPingResults(stats *network.PingStats) {
	// Find max RTT for bar scaling.
	var maxRTT time.Duration
	for _, r := range stats.Results {
		if r.Time > maxRTT {
			maxRTT = r.Time
		}
	}

	// Print individual results with bars.
	for _, r := range stats.Results {
		bar := network.PingBar(r.Time, maxRTT, 20)
		fmt.Printf("  %2d: %6.1fms  %s\n", r.Seq+1, float64(r.Time)/float64(time.Millisecond), bar)
	}

	// Print statistics.
	fmt.Println()
	fmt.Println("Statistics:")
	fmt.Printf("  Sent: %d  Received: %d  Lost: %d (%.0f%%)\n",
		stats.Sent, stats.Received, stats.Lost, stats.LossPercent)
	fmt.Printf("  Min: %.1fms  Avg: %.1fms  Max: %.1fms  StdDev: %.1fms\n",
		float64(stats.MinRTT)/float64(time.Millisecond),
		float64(stats.AvgRTT)/float64(time.Millisecond),
		float64(stats.MaxRTT)/float64(time.Millisecond),
		float64(stats.StdDevRTT)/float64(time.Millisecond))
	fmt.Println()
}
