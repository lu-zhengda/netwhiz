package network

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// NetworkEvent represents a network-related system log event.
type NetworkEvent struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Interface string `json:"interface,omitempty"`
	Detail    string `json:"detail"`
}

// EventService retrieves network-related events from system logs.
type EventService struct {
	runner CmdRunner
}

// NewEventService creates a new EventService.
func NewEventService(runner CmdRunner) *EventService {
	return &EventService{runner: runner}
}

// GetEvents retrieves network events from the last `duration` period.
// duration should be in a format accepted by `log show --last`, e.g. "24h", "1h", "30m".
func (s *EventService) GetEvents(ctx context.Context, duration string) ([]NetworkEvent, error) {
	out, err := s.runner.Run(ctx, "log", "show",
		"--predicate", `subsystem == "com.apple.network" OR subsystem == "com.apple.wifi"`,
		"--style", "compact",
		"--last", duration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read system log: %w", err)
	}
	return ParseNetworkEvents(string(out)), nil
}

// wifiEventPatterns maps regex patterns to event types for WiFi events.
var wifiEventPatterns = []struct {
	pattern   *regexp.Regexp
	eventType string
}{
	{regexp.MustCompile(`(?i)disassociated`), "wifi_disconnect"},
	{regexp.MustCompile(`(?i)associated`), "wifi_connect"},
	{regexp.MustCompile(`(?i)SSID`), "wifi_ssid_change"},
	{regexp.MustCompile(`(?i)roam`), "wifi_roam"},
}

// networkEventPatterns maps regex patterns to event types for network events.
var networkEventPatterns = []struct {
	pattern   *regexp.Regexp
	eventType string
}{
	{regexp.MustCompile(`(?i)address.*changed|new.*address|address.*added`), "ip_change"},
	{regexp.MustCompile(`(?i)DNS.*config|resolver.*config`), "dns_change"},
	{regexp.MustCompile(`(?i)interface.*up|link.*up`), "interface_up"},
	{regexp.MustCompile(`(?i)interface.*down|link.*down`), "interface_down"},
	{regexp.MustCompile(`(?i)VPN.*disconnect|tunnel.*tear`), "vpn_disconnect"},
	{regexp.MustCompile(`(?i)VPN.*connect|tunnel.*establish`), "vpn_connect"},
}

// interfaceRegexp extracts interface names like en0, en1, utun0, etc.
var interfaceRegexp = regexp.MustCompile(`\b(en\d+|utun\d+|lo\d+|bridge\d+|awdl\d+|llw\d+)\b`)

// logLineRegexp parses compact log format lines.
// Example: 2024-01-15 10:30:45.123456-0800  pid  subsystem  message
var logLineRegexp = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\.\d+[\-+]\d{4})\s+(.+)$`)

// ParseNetworkEvents parses log output and extracts network-related events.
func ParseNetworkEvents(output string) []NetworkEvent {
	var events []NetworkEvent

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := logLineRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		timestamp := matches[1]
		message := matches[2]

		event := classifyEvent(message)
		if event == nil {
			continue
		}

		event.Timestamp = timestamp

		// Try to extract interface name.
		if ifMatch := interfaceRegexp.FindString(message); ifMatch != "" {
			event.Interface = ifMatch
		}

		events = append(events, *event)
	}

	return events
}

// classifyEvent determines if a log message is a relevant network event.
func classifyEvent(message string) *NetworkEvent {
	// Check WiFi patterns first.
	for _, p := range wifiEventPatterns {
		if p.pattern.MatchString(message) {
			return &NetworkEvent{
				Type:   p.eventType,
				Detail: summarizeMessage(message),
			}
		}
	}

	// Check network patterns.
	for _, p := range networkEventPatterns {
		if p.pattern.MatchString(message) {
			return &NetworkEvent{
				Type:   p.eventType,
				Detail: summarizeMessage(message),
			}
		}
	}

	return nil
}

// summarizeMessage trims a log message to a reasonable length for display.
func summarizeMessage(message string) string {
	const maxLen = 200
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen-3] + "..."
}
