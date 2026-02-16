package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/lu-zhengda/netwhiz/internal/network"
	"github.com/spf13/cobra"
)

var eventsLast string
var typeFilter string

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Show recent network change events from system log",
	Long:  "Parse system log for network-related events including WiFi disconnections, IP changes, DNS changes, and interface events.",
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		svc := network.NewEventService(runner)

		ctx := context.Background()
		events, err := svc.GetEvents(ctx, eventsLast)
		if err != nil {
			return fmt.Errorf("failed to get network events: %w", err)
		}

		if typeFilter != "" {
			var filtered []network.NetworkEvent
			for _, e := range events {
				if e.Type == typeFilter {
					filtered = append(filtered, e)
				}
			}
			events = filtered
		}

		events = network.DeduplicateEvents(events, 30*time.Second)

		if jsonFlag {
			return printJSON(events)
		}

		printNetworkEvents(events)
		return nil
	},
}

func init() {
	eventsCmd.Flags().StringVar(&eventsLast, "last", "24h", "Time range to search (e.g., 1h, 30m, 24h, 7d)")
	eventsCmd.Flags().StringVar(&typeFilter, "type", "", "Filter events by type (e.g., wifi_disconnect, connection_drop, path_satisfied)")
}

func printNetworkEvents(events []network.NetworkEvent) {
	fmt.Println()
	fmt.Println("Network Events")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")
	fmt.Println()

	if len(events) == 0 {
		fmt.Println("  No network events found in the specified time range.")
		fmt.Println()
		return
	}

	fmt.Printf("  Found %d event(s):\n\n", len(events))

	for _, event := range events {
		iface := ""
		if event.Interface != "" {
			iface = fmt.Sprintf(" [%s]", event.Interface)
		}
		eventType := event.Type
		if event.Count > 1 {
			eventType = fmt.Sprintf("%s (x%d)", event.Type, event.Count)
		}
		fmt.Printf("  %s  %-20s%s\n", event.Timestamp, eventType, iface)
		fmt.Printf("    %s\n\n", event.Detail)
	}
}
