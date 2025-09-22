# Calendar Widget for Waybar

A Go-based calendar widget for waybar that displays your Microsoft 365 calendar with visual indicators and click-to-join functionality for Teams meetings.

## Features

- 🔴 **Smart Visual Indicators**: Shows current, urgent (≤5min), soon (≤15min), upcoming, and past meetings
- 📅 **Microsoft 365 Integration**: Full calendar access using Microsoft Graph CalendarView API
- 🔗 **Teams Meeting Support**: Direct Teams app integration with automatic Teams link detection
- 💡 **Rich Tooltips**: Shows today's full schedule + upcoming events
- 🔄 **Smart Authentication**: Browser-based login with automatic token refresh - no app registration required!
- 👆 **Intelligent Clicks**: Auto-opens current/urgent meetings, handles auth errors gracefully
- ⚡ **Lightweight & Fast**: Built with Go for optimal performance

## Installation

### Prerequisites

- Go 1.21 or later
- Linux with waybar and a web browser

### Build from source

```bash
git clone https://github.com/magudb/waybar-calendar.git
cd calendar-widget
go build -o calendar-widget
sudo cp calendar-widget /usr/local/bin/
```

## Setup

### Simple One-Command Setup

```bash
calendar-widget setup
```

That's it! The widget uses Microsoft's public client authentication - **no Azure app registration required!**

The setup will:
1. 🌐 Open your browser for Microsoft login
2. 🔐 Securely cache your credentials locally
3. ✅ Test calendar access
4. 🎉 Ready to use!

### What You Get

- **Personal & Work Accounts**: Supports both microsoft.com and organizational accounts
- **Automatic Token Refresh**: No need to re-authenticate frequently
- **Secure Local Storage**: Tokens stored in `~/.config/calendar-widget/`

## Waybar Configuration

Add this to your waybar config (`~/.config/waybar/config.jsonc`):

```json
{
    "modules-center": ["custom/calendar-widget"],
    "custom/calendar-widget": {
        "exec": "calendar-widget waybar",
        "return-type": "json",
        "interval": 60,
        "on-click": "calendar-widget click",
        "tooltip": true,
        "exec-tooltip": "calendar-widget tooltip",
        "signal": 8
    }
}
```

### Advanced Click Handling

The `calendar-widget click` command intelligently handles:
- **🔐 Auth Required** → Automatically runs `calendar-widget reauth`
- **🟢 Current Meeting** → Opens Teams/browser link directly
- **🔴 Urgent Meeting** → Opens Teams/browser link directly
- **📅 Other Times** → Opens calendar widget interface

### Waybar CSS Styling

Add to your waybar CSS (`~/.config/waybar/style.css`):

```css
/* Calendar Widget Styles */
#custom-calendar-widget {
    padding: 0 10px;
    margin: 0 5px;
    border-radius: 5px;
    font-weight: bold;
    transition: all 0.3s ease;
}

#custom-calendar-widget.urgent {
    background-color: #ff4444;
    color: #ffffff;
    animation: pulse 1s infinite;
}

#custom-calendar-widget.soon {
    background-color: #ffaa00;
    color: #000000;
}

#custom-calendar-widget.upcoming {
    background-color: #4488ff;
    color: #ffffff;
}

#custom-calendar-widget.current {
    background-color: #44ff44;
    color: #000000;
    animation: pulse 2s infinite;
}

#custom-calendar-widget.past {
    background-color: #666666;
    color: #cccccc;
}

#custom-calendar-widget.no-meeting {
    background-color: transparent;
    color: #888888;
}

#custom-calendar-widget.error {
    background-color: #ff0000;
    color: #ffffff;
}

/* Pulse animation for urgent and current meetings */
@keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.7; }
    100% { opacity: 1; }
}
```

## Usage

### Available Commands

```bash
# Initial setup (run once)
calendar-widget setup

# Run waybar integration (called by waybar)
calendar-widget waybar

# Smart click handler (called by waybar on-click)
calendar-widget click

# Show detailed tooltip (called by waybar exec-tooltip)
calendar-widget tooltip

# Run interactive widget (TUI interface)
calendar-widget widget

# Re-authenticate (clear tokens and login again)
calendar-widget reauth

# Clear all credentials and exit
calendar-widget logout

# Debug calendar access and events
calendar-widget debug
```

### Visual Status Indicators

| Status | Icon | Color | Description |
|--------|------|-------|-------------|
| 🟢 Current | Green | Meeting happening now |
| 🔴 Urgent | Red | Meeting starts ≤5 minutes |
| 🟡 Soon | Yellow | Meeting starts ≤15 minutes |
| 🔵 Upcoming | Blue | Meeting starts >15 minutes |
| ⚫ Past | Gray | Meeting already ended |

### Teams Integration

- **🔗 Automatic Detection**: Uses Microsoft Graph `onlineMeeting` field
- **[T] Indicator**: Teams meetings show "[T]" prefix in widget text
- **Direct Launch**: Click opens Teams app directly, not browser
- **Fallback Support**: Detects Teams links in body text for edge cases

## How It Works

### Microsoft Graph CalendarView API
The widget uses Microsoft Graph's CalendarView endpoint for accurate calendar data:
```
GET /v1.0/users/{user}/calendarView?StartDateTime={start}&endDateTime={end}
```

This provides:
- ✅ **Accurate Today's Events**: Proper date range filtering
- ✅ **Teams Meeting Data**: `onlineMeeting.joinUrl` field
- ✅ **Timezone Handling**: Automatic local timezone conversion
- ✅ **Recurring Events**: Expanded recurring series

### Smart Tooltip System
- **Today's Schedule**: Shows all events for current day
- **Upcoming Events**: Shows next 5 events with smart date formatting
- **Status Indicators**: Color-coded by urgency/timing
- **Teams Detection**: Clear "(Teams)" indicators

## Configuration Files

- **Config**: `~/.config/calendar-widget/config.json`
- **Tokens**: `~/.config/calendar-widget/token.json` (automatically managed)

## Troubleshooting

### Authentication Issues

```bash
# Re-authenticate (clears tokens and re-login)
calendar-widget reauth

# Check authentication status
calendar-widget debug

# Complete fresh start
calendar-widget logout && calendar-widget setup
```

### Waybar Integration Issues

1. **Check Waybar Logs**: `journalctl -u waybar`
2. **Test Manually**: `calendar-widget waybar`
3. **Verify JSON**: Should output valid JSON with `text`, `class`, `tooltip`
4. **Check Permissions**: `chmod +x /usr/local/bin/calendar-widget`

### Common Issues

| Issue | Solution |
|-------|----------|
| "Auth Required" in waybar | Click the widget or run `calendar-widget reauth` |
| No events showing | Run `calendar-widget debug` to check API response |
| Teams links not working | Ensure Teams app is installed and configured |
| Widget not updating | Check waybar interval setting (60s recommended) |

## Example Output

### Waybar Widget Display
```
🟢 Office                    # Current meeting
🔴 Daily Standup (in 3m)     # Urgent meeting
🟡 [T] Team Sync (in 12m)    # Teams meeting soon
🔵 Project Review (in 45m)   # Upcoming meeting
```

### Tooltip Content
```
📅 Today's Schedule:

🟢 02:00-02:00 Office
⚫ 08:00-10:00 Getting ready
⚫ 09:00-09:15 Update dashboard
⚫ 10:00-10:30 Daily standups (Teams)
🔵 12:00-13:00 Lunch and ToDos

🔮 Upcoming Events:

🔵 Wed 24/9 13:30  Legal advice
🔵 Thu 25/9 10:00  Team meeting (Teams)
🔵 Thu 25/9 10:30  Workshop (Teams)
```

## Development

### Building

```bash
go build -o calendar-widget
```

### Key Dependencies

- **[Microsoft Graph SDK Go](https://github.com/microsoftgraph/msgraph-sdk-go)** - Microsoft 365 API access
- **[Azure Identity Go](https://github.com/Azure/azure-sdk-for-go/sdk/azidentity)** - Authentication
- **[Cobra](https://github.com/spf13/cobra)** - CLI framework
- **[Bubbletea](https://github.com/charmbracelet/bubbletea)** - TUI interface
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling

### Project Structure

```
calendar-widget/
├── cmd/                    # CLI commands
│   ├── root.go
│   ├── setup.go           # Authentication setup
│   ├── waybar.go          # Waybar integration
│   ├── click.go           # Smart click handler
│   └── ...
├── internal/
│   ├── auth/              # Authentication logic
│   ├── calendar/          # Microsoft Graph API
│   └── widget/            # UI components
└── main.go
```

## License

MIT License - see LICENSE file for details.