package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

var speedCmd = &cobra.Command{
	Use:   "speed",
	Short: "Run a speed test",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewSpeedService(runner)

		ctx := context.Background()

		fmt.Println()
		fmt.Println("Speed Test")
		fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
		fmt.Println()
		fmt.Println("  Testing download speed (Cloudflare)...")

		result, err := svc.RunSpeedTest(ctx)
		if err != nil {
			return fmt.Errorf("failed to run speed test: %w", err)
		}

		fmt.Println()
		fmt.Printf("  Download:  %s\n", network.FormatSpeed(result.DownloadMbps))
		fmt.Printf("  Latency:   %.0fms\n", float64(result.Latency)/float64(time.Millisecond))
		fmt.Printf("  Server:    %s\n", result.Server)
		fmt.Printf("  Time:      %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Println()

		return nil
	},
}
