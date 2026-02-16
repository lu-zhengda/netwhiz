package network

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// NetworkEvent represents a network-related system log event.
type NetworkEvent struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Interface string `json:"interface,omitempty"`
	Detail    string `json:"detail"`
	Count     int    `json:"count,omitempty"`
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
	{regexp.MustCompile(`(?i)disassociat`), "wifi_disconnect"},
	{regexp.MustCompile(`(?i)\bassociat`), "wifi_connect"},
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
	{regexp.MustCompile(`(?i)path:unsatisfied|path became unsatisfied`), "path_unsatisfied"},
	{regexp.MustCompile(`(?i)path:satisfied[^_]|path became satisfied`), "path_satisfied"},
	{regexp.MustCompile(`(?i)flow:disconnect`), "connection_drop"},
}

// interfaceRegexp extracts interface names like en0, en1, utun0, etc.
var interfaceRegexp = regexp.MustCompile(`\b(en\d+|utun\d+|lo\d+|bridge\d+|awdl\d+|llw\d+)\b`)

// logLineRegexp parses compact log format lines.
// Example: 2024-01-15 10:30:45.123 Df wifid[234:1234] [com.apple.wifi:manager] message
var logLineRegexp = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\.\d+(?:[\-+]\d{4})?)\s+(.+)$`)

// logPrefixRegexp matches the compact log prefix: "Ty process[PID:TID] [subsystem:category] "
var logPrefixRegexp = regexp.MustCompile(`^\w{1,3}\s+\S+\[[^\]]+\]\s+\[[^\]]+\]\s+`)

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

// summarizeMessage strips the compact log prefix and trims to a reasonable length.
func summarizeMessage(message string) string {
	msg := logPrefixRegexp.ReplaceAllString(message, "")

	const maxLen = 200
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}

// DeduplicateEvents collapses consecutive events of the same type
// within a time window into a single event with a count.
func DeduplicateEvents(events []NetworkEvent, window time.Duration) []NetworkEvent {
	if len(events) == 0 {
		return nil
	}

	const tsLayout = "2006-01-02 15:04:05"

	// parseTS extracts just the date+time portion for comparison.
	parseTS := func(ts string) (time.Time, bool) {
		// Timestamps may include fractional seconds and timezone offsets;
		// truncate to "2006-01-02 15:04:05" for grouping.
		if len(ts) >= len(tsLayout) {
			t, err := time.Parse(tsLayout, ts[:len(tsLayout)])
			if err == nil {
				return t, true
			}
		}
		return time.Time{}, false
	}

	var result []NetworkEvent

	current := events[0]
	current.Count = 1
	groupStart, _ := parseTS(current.Timestamp)

	for i := 1; i < len(events); i++ {
		ev := events[i]
		evTime, ok := parseTS(ev.Timestamp)

		if ev.Type == current.Type && ok && evTime.Sub(groupStart) <= window {
			current.Count++
		} else {
			result = append(result, current)
			current = ev
			current.Count = 1
			groupStart = evTime
		}
	}

	result = append(result, current)
	return result
}
