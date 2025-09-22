package cmd

import (
	"calendar-widget/internal/widget"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var tooltipCmd = &cobra.Command{
	Use:   "tooltip",
	Short: "Show tooltip with full day schedule",
	Long:  `Display a tooltip showing the full day's calendar events.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runTooltip(); err != nil {
			fmt.Printf("Tooltip failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runTooltip() error {
	w, err := widget.NewWidget(&widget.Config{
		Debug: debug,
	})
	if err != nil {
		return fmt.Errorf("failed to create widget: %w", err)
	}

	return w.ShowTooltip()
}
