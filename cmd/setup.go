package cmd

import (
	"calendar-widget/internal/auth"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Microsoft 365 authentication",
	Long: `Setup authentication for Microsoft 365 calendar access.
This will authenticate you with Microsoft using a standard login flow - no app registration required!`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSetup(); err != nil {
			fmt.Printf("Setup failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runSetup() error {
	fmt.Println("Calendar Widget Setup")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("Welcome! This setup will authenticate you with Microsoft 365 to access your calendar.")
	fmt.Println("No app registration required - we'll use Microsoft's standard authentication flow.")
	fmt.Println()
	fmt.Println("This widget can access:")
	fmt.Println("• Your calendar events (read-only)")
	fmt.Println("• Your basic profile information")
	fmt.Println()
	fmt.Println("Your credentials will be securely cached locally for future use.")
	fmt.Println()

	// Create default public client config
	config := &auth.Config{
		ClientID:    auth.PublicClientID,
		TenantID:    auth.CommonTenant,
		RedirectURI: auth.RedirectURI,
		UsePublic:   true,
	}

	// Save the default config
	if err := auth.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Starting authentication process...")
	fmt.Println("Your default browser will open for Microsoft login.")
	fmt.Println("Please complete the authentication in your browser.")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	_, err := auth.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Authentication successful!")
	fmt.Println("✅ Credentials cached for future use")
	fmt.Println()
	fmt.Println("Setup complete! You can now use the calendar widget.")
	fmt.Println("Try running: calendar-widget")

	return nil
}
