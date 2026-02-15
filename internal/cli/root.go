package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zhengda-lu/netwhiz/internal/network"
	"github.com/zhengda-lu/netwhiz/internal/tui"
)

var (
	// Set via ldflags at build time.
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "netwhiz",
	Short:   "A network diagnostics toolkit for macOS",
	Long:    "netwhiz is a Swiss-army-knife for macOS network diagnostics.\nLaunch without subcommands for interactive TUI mode.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := &network.RealCmdRunner{}
		p := tea.NewProgram(tui.New(runner, version), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("netwhiz %s\n", version))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(dnsCmd)
	rootCmd.AddCommand(wifiCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(speedCmd)
	rootCmd.AddCommand(scanCmd)
}
