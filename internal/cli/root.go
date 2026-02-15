package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/lu-zhengda/netwhiz/internal/network"
	"github.com/lu-zhengda/netwhiz/internal/tui"
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
		if shell, _ := cmd.Flags().GetString("generate-completion"); shell != "" {
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
			}
		}
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
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")
	rootCmd.SetVersionTemplate(fmt.Sprintf("netwhiz %s\n", version))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().String("generate-completion", "", "Generate shell completion (bash, zsh, fish)")
	rootCmd.Flags().MarkHidden("generate-completion")
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(dnsCmd)
	rootCmd.AddCommand(wifiCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(speedCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(diagnoseCmd)
	rootCmd.AddCommand(vpnCmd)
	rootCmd.AddCommand(eventsCmd)
}
