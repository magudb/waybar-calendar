package cmd

import (
	"calendar-widget/internal/auth"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var reauthCmd = &cobra.Command{
	Use:   "reauth",
	Short: "Clear tokens and re-authenticate",
	Long:  `Clear stored tokens and re-authenticate with Microsoft 365.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runReauth(); err != nil {
			fmt.Printf("Re-authentication failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runReauth() error {
	// Clear existing tokens
	if err := auth.ClearTokens(); err != nil {
		fmt.Printf("Warning: failed to clear tokens: %v\n", err)
	}

	fmt.Println("🔄 Re-authenticating...")
	fmt.Println("Starting fresh authentication process...")

	// Run setup again
	return runSetup()
}

func init() {
	rootCmd.AddCommand(reauthCmd)
}
