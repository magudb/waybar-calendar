package calendar

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"calendar-widget/internal/auth"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go-core/authentication"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

type Event struct {
	Subject   string
	Start     time.Time
	End       time.Time
	Location  string
	WebLink   string
	TeamsLink string
	IsTeams   bool
	Organizer string
	Attendees []string
	Body      string
}

type CalendarService struct {
	client *msgraphsdk.GraphServiceClient
}

func NewCalendarService() (*CalendarService, error) {
	return NewCalendarServiceWithOptions(true)
}

func NewCalendarServiceWithOptions(allowInteractive bool) (*CalendarService, error) {
	// Create a custom credential that respects interactive mode
	credential := &nonInteractiveCredential{allowInteractive: allowInteractive}

	authProvider, err := authentication.NewAzureIdentityAuthenticationProviderWithScopes(credential, []string{
		"https://graph.microsoft.com/Calendars.Read",
		"https://graph.microsoft.com/User.Read",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	client := msgraphsdk.NewGraphServiceClient(adapter)

	return &CalendarService{client: client}, nil
}

// nonInteractiveCredential wraps the authentication to control interactive behavior
type nonInteractiveCredential struct {
	allowInteractive bool
}

func (nic *nonInteractiveCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return auth.GetAccessTokenWithOptions(ctx, nic.allowInteractive)
}

func (cs *CalendarService) GetTodaysEvents(ctx context.Context) ([]Event, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Use CalendarView with proper date range
	startStr := startOfDay.UTC().Format("2006-01-02T15:04:05.000Z")
	endStr := endOfDay.UTC().Format("2006-01-02T15:04:05.000Z")

	return cs.getEventsWithCalendarView(ctx, startStr, endStr)
}

func (cs *CalendarService) GetUpcomingEvents(ctx context.Context) ([]Event, error) {
	now := time.Now()
	// Get events from now until 7 days from now
	endTime := now.Add(7 * 24 * time.Hour)

	// Use CalendarView with proper date range
	nowStr := now.UTC().Format("2006-01-02T15:04:05.000Z")
	endStr := endTime.UTC().Format("2006-01-02T15:04:05.000Z")

	return cs.getEventsWithCalendarView(ctx, nowStr, endStr)
}

func (cs *CalendarService) getEventsWithCalendarView(ctx context.Context, startDateTime, endDateTime string) ([]Event, error) {
	requestConfiguration := &users.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarViewRequestBuilderGetQueryParameters{
			StartDateTime: &startDateTime,
			EndDateTime:   &endDateTime,
			Orderby:       []string{"start/dateTime"},
			Select:        []string{"subject", "start", "end", "location", "webLink", "body", "organizer", "attendees", "onlineMeeting"},
			Top:           intPtr(50),
		},
	}

	events, err := cs.client.Me().CalendarView().Get(ctx, requestConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar view: %w", err)
	}

	var result []Event
	for _, event := range events.GetValue() {
		e := Event{
			Subject:  getStringValue(event.GetSubject()),
			Location: getStringValue(event.GetLocation().GetDisplayName()),
			WebLink:  getStringValue(event.GetWebLink()),
			Body:     getStringValue(event.GetBody().GetContent()),
		}

		if event.GetStart() != nil && event.GetStart().GetDateTime() != nil {
			startStr := getStringValue(event.GetStart().GetDateTime())
			e.Start = parseMicrosoftDateTime(startStr)
		}
		if event.GetEnd() != nil && event.GetEnd().GetDateTime() != nil {
			endStr := getStringValue(event.GetEnd().GetDateTime())
			e.End = parseMicrosoftDateTime(endStr)
		}

		if event.GetOrganizer() != nil && event.GetOrganizer().GetEmailAddress() != nil {
			e.Organizer = getStringValue(event.GetOrganizer().GetEmailAddress().GetName())
		}

		for _, attendee := range event.GetAttendees() {
			if attendee.GetEmailAddress() != nil {
				e.Attendees = append(e.Attendees, getStringValue(attendee.GetEmailAddress().GetName()))
			}
		}

		// Use onlineMeeting field for Teams meetings
		if event.GetOnlineMeeting() != nil {
			e.IsTeams = true
			if event.GetOnlineMeeting().GetJoinUrl() != nil {
				e.TeamsLink = getStringValue(event.GetOnlineMeeting().GetJoinUrl())
			}
		} else {
			// Fallback to body/location parsing for non-standard meeting links
			e.TeamsLink, e.IsTeams = extractTeamsLink(e.Body, e.Location)
		}

		result = append(result, e)
	}

	return result, nil
}

func (cs *CalendarService) GetNextMeeting(ctx context.Context) (*Event, error) {
	events, err := cs.GetUpcomingEvents(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, event := range events {
		if event.Start.After(now) || (event.Start.Before(now) && event.End.After(now)) {
			return &event, nil
		}
	}

	return nil, nil
}

func extractTeamsLink(body, location string) (string, bool) {
	// Multiple Teams URL patterns to look for
	teamsPatterns := []string{
		`https://teams\.microsoft\.com/l/meetup-join/[^\s<>"']+`,
		`https://teams\.live\.com/meet/[^\s<>"']+`,
		`https://[a-zA-Z0-9-]+\.teams\.microsoft\.com/[^\s<>"']+`,
	}

	content := body + " " + location

	// Try each Teams URL pattern
	for _, pattern := range teamsPatterns {
		teamsRegex := regexp.MustCompile(pattern)
		if match := teamsRegex.FindString(content); match != "" {
			// Clean up the URL (remove trailing punctuation)
			cleanURL := strings.TrimRight(match, ".,:;!?")
			return cleanURL, true
		}
	}

	// Look for Teams meeting indicators
	teamsIndicators := []string{
		"Microsoft Teams Meeting",
		"Teams Meeting",
		"Join Microsoft Teams Meeting",
		"Microsoft Teams-møde", // Danish
		"Teams-møde",           // Danish
	}

	contentLower := strings.ToLower(content)

	for _, indicator := range teamsIndicators {
		if strings.Contains(contentLower, strings.ToLower(indicator)) {
			// Extract any HTTPS URL from the content
			urlRegex := regexp.MustCompile(`https://[^\s<>"']+`)
			matches := urlRegex.FindAllString(content, -1)
			for _, match := range matches {
				cleanURL := strings.TrimRight(match, ".,:;!?")
				if u, err := url.Parse(cleanURL); err == nil && u.Host != "" {
					return cleanURL, true
				}
			}
			// Found Teams indicator but no usable URL
			return "", true
		}
	}

	return "", false
}

func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func intPtr(i int32) *int32 {
	return &i
}

func parseMicrosoftDateTime(dateTimeStr string) time.Time {
	if dateTimeStr == "" {
		return time.Time{}
	}

	// Microsoft Graph datetime formats to try
	formats := []string{
		"2006-01-02T15:04:05.0000000", // Microsoft's .NET format
		"2006-01-02T15:04:05",         // ISO without fractional seconds
		time.RFC3339,                  // Standard RFC3339
		"2006-01-02T15:04:05Z",        // UTC format
		"2006-01-02T15:04:05.000Z",    // UTC with milliseconds
	}

	for _, format := range formats {
		if parsedTime, err := time.Parse(format, dateTimeStr); err == nil {
			// Microsoft Graph returns times in the user's timezone, but without timezone info
			// We'll assume local timezone if no timezone is specified
			if parsedTime.Location() == time.UTC && !strings.HasSuffix(dateTimeStr, "Z") {
				// Convert to local timezone
				return parsedTime.In(time.Local)
			}
			return parsedTime
		}
	}

	// If all formats fail, return zero time
	return time.Time{}
}

func (e *Event) GetTimeUntil() time.Duration {
	return time.Until(e.Start)
}

func (e *Event) GetStatus() string {
	now := time.Now()
	if now.After(e.End) {
		return "past"
	}
	if now.After(e.Start) && now.Before(e.End) {
		return "current"
	}

	timeUntil := time.Until(e.Start)
	if timeUntil <= 5*time.Minute {
		return "urgent"
	}
	if timeUntil <= 15*time.Minute {
		return "soon"
	}
	return "upcoming"
}
