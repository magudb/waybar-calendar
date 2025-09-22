package cmd

import (
	"calendar-widget/internal/calendar"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug calendar access",
	Long:  `Debug command to test calendar access and show detailed information.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDebug(); err != nil {
			fmt.Printf("Debug failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runDebug() error {
	fmt.Println("🔍 Debug Calendar Access")
	fmt.Println("========================")

	calendarService, err := calendar.NewCalendarService()
	if err != nil {
		return fmt.Errorf("failed to create calendar service: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("📅 Current time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("🌍 Timezone: %s\n", time.Now().Location())
	fmt.Println()

	fmt.Println("📋 Getting today's events...")
	todaysEvents, err := calendarService.GetTodaysEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get today's events: %w", err)
	}

	fmt.Printf("📊 Found %d today's events\n", len(todaysEvents))
	fmt.Println()

	fmt.Println("📋 Getting upcoming events...")
	upcomingEvents, err := calendarService.GetUpcomingEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get upcoming events: %w", err)
	}

	fmt.Printf("📊 Found %d upcoming events\n", len(upcomingEvents))
	fmt.Println()

	// Show today's events first
	events := todaysEvents
	if len(todaysEvents) == 0 && len(upcomingEvents) > 0 {
		fmt.Println("📌 No events today, showing upcoming events instead:")
		events = upcomingEvents
	}

	if len(events) == 0 {
		fmt.Println("❌ No events found")
		fmt.Println("This could be because:")
		fmt.Println("  • Events are in a different timezone")
		fmt.Println("  • Events are in a different calendar")
		fmt.Println("  • Query filter is too restrictive")
		return nil
	}

	for i, event := range events {
		fmt.Printf("📅 Event %d:\n", i+1)
		fmt.Printf("  📝 Subject: %s\n", event.Subject)
		fmt.Printf("  🕐 Start: %s\n", event.Start.Format(time.RFC3339))
		fmt.Printf("  🕐 End: %s\n", event.End.Format(time.RFC3339))
		fmt.Printf("  📍 Location: %s\n", event.Location)
		fmt.Printf("  🔗 Teams: %t\n", event.IsTeams)
		if event.TeamsLink != "" {
			fmt.Printf("  🔗 Teams Link: %s\n", event.TeamsLink)
		}
		fmt.Printf("  🌐 Web Link: %s\n", event.WebLink)
		fmt.Printf("  📊 Status: %s\n", event.GetStatus())

		// Show only upcoming events to reduce noise
		if event.GetStatus() != "past" {
			fmt.Printf("  ⏰ Time until: %v\n", event.GetTimeUntil())
		}
		fmt.Println()

		// Limit to first 5 events for readability
		if i >= 4 {
			fmt.Printf("... and %d more events\n\n", len(events)-5)
			break
		}
	}

	fmt.Println("🔍 Getting next meeting...")
	nextMeeting, err := calendarService.GetNextMeeting(ctx)
	if err != nil {
		return fmt.Errorf("failed to get next meeting: %w", err)
	}

	if nextMeeting == nil {
		fmt.Println("❌ No next meeting found")
	} else {
		fmt.Printf("📅 Next meeting: %s\n", nextMeeting.Subject)
		fmt.Printf("🕐 Starts: %s\n", nextMeeting.Start.Format(time.RFC3339))
		fmt.Printf("📊 Status: %s\n", nextMeeting.GetStatus())
		timeUntil := nextMeeting.GetTimeUntil()
		if timeUntil > 0 {
			fmt.Printf("⏰ Time until: %v\n", timeUntil)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
