package cmd

import (
	"calendar-widget/internal/widget"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	refresh int
	compact bool
)

var widgetCmd = &cobra.Command{
	Use:   "widget",
	Short: "Run the calendar widget",
	Long:  `Run the calendar widget that displays your next meeting.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runWidget(); err != nil {
			fmt.Printf("Widget failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runWidget() error {
	w, err := widget.NewWidget(&widget.Config{
		RefreshInterval: refresh,
		Compact:         compact,
		Debug:           debug,
	})
	if err != nil {
		return fmt.Errorf("failed to create widget: %w", err)
	}

	return w.Run()
}

func init() {
	widgetCmd.Flags().IntVar(&refresh, "refresh", 60, "refresh interval in seconds")
	widgetCmd.Flags().BoolVar(&compact, "compact", false, "use compact display mode")
}
