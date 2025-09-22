package cmd

import (
	"calendar-widget/internal/auth"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear stored authentication tokens",
	Long:  `Logout from Microsoft 365 and clear any stored authentication tokens and configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runLogout(); err != nil {
			fmt.Printf("Logout failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runLogout() error {
	fmt.Println("Logging out...")

	// Remove token file
	tokenPath := auth.GetTokenPath()
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}

	// Remove config file
	configPath := auth.GetConfigPath()
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	fmt.Println("✅ Successfully logged out!")
	fmt.Println("All stored authentication data has been cleared.")
	fmt.Println()
	fmt.Println("To use the calendar widget again, run: calendar-widget setup")

	return nil
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
