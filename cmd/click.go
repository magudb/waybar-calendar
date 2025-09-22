package cmd

import (
	"calendar-widget/internal/calendar"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var clickCmd = &cobra.Command{
	Use:   "click",
	Short: "Handle calendar widget clicks intelligently",
	Long:  `Handle clicks on the calendar widget. If authentication is required, run reauth. Otherwise, open the current meeting.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runClick(); err != nil {
			fmt.Printf("Click handler failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runClick() error {
	// First check if authentication is working by trying to create a calendar service
	calendarService, err := calendar.NewCalendarServiceWithOptions(false) // Non-interactive
	if err != nil {
		if isAuthError(err) {
			fmt.Println("Authentication required, running reauth...")
			return runReauth()
		}
		// If there's an error but not auth-related, just open the widget
		return runWidget()
	}

	// Try to get upcoming events to see if there's a current/urgent meeting
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	upcomingEvents, err := calendarService.GetUpcomingEvents(ctx)
	if err != nil {
		if isAuthError(err) {
			fmt.Println("Authentication required, running reauth...")
			return runReauth()
		}
		// If there's an error but not auth-related, just open the widget
		return runWidget()
	}

	// Look for current or urgent meetings to open
	for _, event := range upcomingEvents {
		status := event.GetStatus()
		if status == "current" || status == "urgent" {
			// Open this meeting
			if event.IsTeams && event.TeamsLink != "" {
				return openMeetingLink(event.TeamsLink)
			} else if event.WebLink != "" {
				return openMeetingLink(event.WebLink)
			}
		}
	}

	// No current/urgent meetings, just run the regular widget
	return runWidget()
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "token") ||
		strings.Contains(errStr, "login") ||
		strings.Contains(errStr, "unauthorized")
}

func openMeetingLink(url string) error {
	// Use the same logic as the widget's openMeeting function
	var cmd string
	switch {
	case strings.Contains(url, "teams.microsoft.com"):
		// Try to open in Teams app first, fallback to browser
		cmd = fmt.Sprintf(`sh -c 'xdg-open "msteams://" 2>/dev/null && sleep 1 && xdg-open "%s" || xdg-open "%s"'`, url, url)
	default:
		cmd = fmt.Sprintf(`xdg-open "%s"`, url)
	}

	return runBashCommand(cmd)
}

func runBashCommand(command string) error {
	// Execute the command using shell
	exec := exec.Command("sh", "-c", command)
	return exec.Run()
}

func init() {
	rootCmd.AddCommand(clickCmd)
}
