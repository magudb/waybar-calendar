package cmd

import (
	"calendar-widget/internal/calendar"
	"calendar-widget/internal/widget"
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
	// First, check what's the current status by running waybar once
	_, err := widget.NewWidgetWithOptions(&widget.Config{
		RefreshInterval: 60,
		Compact:         true,
		Debug:           debug,
	}, false) // Start non-interactive
	if err != nil {
		fmt.Printf("Failed to create widget: %v\n", err)
		return runReauth()
	}

	// Capture output to check for "Auth Required"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to get upcoming events to see what the status is
	calendarService, err := calendar.NewCalendarServiceWithOptions(false)
	if err != nil {
		if isAuthError(err) {
			fmt.Println("Authentication required, forcing token refresh...")
			return runClickWithForceRefresh()
		}
		return runWidget()
	}

	upcomingEvents, err := calendarService.GetUpcomingEvents(ctx)
	if err != nil {
		if isAuthError(err) {
			fmt.Println("Authentication required, forcing token refresh...")
			return runClickWithForceRefresh()
		}
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

func runClickWithForceRefresh() error {
	// Create widget with force refresh
	_, err := widget.NewWidgetWithOptions(&widget.Config{
		RefreshInterval: 60,
		Compact:         true,
		Debug:           debug,
	}, true) // Allow interactive for force refresh
	if err != nil {
		fmt.Printf("Failed to create widget with refresh: %v\n", err)
		return runReauth()
	}

	// Try with force refresh
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	calendarService, err := calendar.NewCalendarServiceWithRefresh(true, true) // Interactive + force refresh
	if err != nil {
		fmt.Printf("Force refresh failed: %v\n", err)
		return runReauth()
	}

	upcomingEvents, err := calendarService.GetUpcomingEvents(ctx)
	if err != nil {
		if isAuthError(err) {
			fmt.Printf("Force refresh still failed with auth error: %v\n", err)
			return runReauth()
		}
		fmt.Printf("Force refresh failed with error: %v\n", err)
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

	// Successfully refreshed but no urgent meetings - just run widget
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
