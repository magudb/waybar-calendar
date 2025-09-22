package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	debug      bool
)

var rootCmd = &cobra.Command{
	Use:   "calendar-widget",
	Short: "A calendar widget for waybar",
	Long: `A calendar widget for waybar that shows your next Microsoft 365 meeting
with visual indicators for urgency and click-to-join functionality for Teams meetings.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Run the widget by default
		widgetCmd.Run(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.config/calendar-widget/config.json)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")

	rootCmd.AddCommand(widgetCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(tooltipCmd)
}
