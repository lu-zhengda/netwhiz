package network

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestParseNetworkEvents(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantTypes []string
	}{
		{
			name: "wifi disconnect event",
			input: `2024-01-15 10:30:45.123 Df wifid[234:1234] [com.apple.wifi:manager] WiFi interface en0 disassociated from network
`,
			wantCount: 1,
			wantTypes: []string{"wifi_disconnect"},
		},
		{
			name: "wifi connect event",
			input: `2024-01-15 10:31:00.123 Df wifid[234:1234] [com.apple.wifi:manager] WiFi interface en0 associated with SSID MyNetwork
`,
			wantCount: 1,
			wantTypes: []string{"wifi_connect"},
		},
		{
			name: "ip address change",
			input: `2024-01-15 10:32:00.123 Df configd[100:5678] [com.apple.network:config] address changed for en0: 192.168.1.100
`,
			wantCount: 1,
			wantTypes: []string{"ip_change"},
		},
		{
			name: "dns configuration change",
			input: `2024-01-15 10:33:00.123 Df mDNSResponder[345:9abc] [com.apple.network:dns] DNS config updated for resolver
`,
			wantCount: 1,
			wantTypes: []string{"dns_change"},
		},
		{
			name: "interface up and down",
			input: `2024-01-15 10:34:00.123 Df configd[100:def0] [com.apple.network:config] interface down: en0
2024-01-15 10:34:05.123 Df configd[100:def0] [com.apple.network:config] interface up: en0
`,
			wantCount: 2,
			wantTypes: []string{"interface_down", "interface_up"},
		},
		{
			name: "multiple mixed events",
			input: `2024-01-15 10:30:00.123 Df wifid[234:1111] [com.apple.wifi:manager] WiFi disassociated from network on en0
2024-01-15 10:30:05.123 Df configd[100:2222] [com.apple.network:config] Some unrelated log message about networking
2024-01-15 10:30:10.123 Df wifid[234:3333] [com.apple.wifi:manager] WiFi associated with SSID OtherNetwork on en0
2024-01-15 10:30:15.123 Df mDNSResponder[345:4444] [com.apple.network:dns] DNS config changed
`,
			wantCount: 3,
			wantTypes: []string{"wifi_disconnect", "wifi_connect", "dns_change"},
		},
		{
			name:      "empty output",
			input:     "",
			wantCount: 0,
			wantTypes: nil,
		},
		{
			name: "no matching events",
			input: `2024-01-15 10:30:00.123 Df configd[100:1111] [com.apple.network:config] routine heartbeat check
2024-01-15 10:31:00.123 Df configd[100:2222] [com.apple.network:config] connection pool status: ok
`,
			wantCount: 0,
			wantTypes: nil,
		},
		{
			name: "malformed lines are skipped",
			input: `This is not a log line
2024-01-15 10:30:00.123 Df wifid[234:1111] [com.apple.wifi:manager] WiFi disassociated
Another bad line
`,
			wantCount: 1,
			wantTypes: []string{"wifi_disconnect"},
		},
		{
			name: "path unsatisfied event",
			input: `2024-01-15 10:30:00.123 Df trustd[554:1111] [com.apple.network:connection] [C5 failed parent-flow] event: path:unsatisfied @30.382s
`,
			wantCount: 1,
			wantTypes: []string{"path_unsatisfied"},
		},
		{
			name: "path satisfied event",
			input: `2024-01-15 10:32:00.123 Df trustd[554:3333] [com.apple.network:connection] [C5 ready parent-flow] event: path:satisfied @0.165s
`,
			wantCount: 1,
			wantTypes: []string{"path_satisfied"},
		},
		{
			name: "flow disconnect event",
			input: `2024-01-15 10:30:00.123 Df trustd[554:1111] [com.apple.network:connection] [C5 failed parent-flow] event: flow:disconnect @30.382s
`,
			wantCount: 1,
			wantTypes: []string{"connection_drop"},
		},
		{
			name: "path became unsatisfied",
			input: `2024-01-15 10:31:00.123 Df mDNSResponder[345:2222] [com.apple.network:path] path became unsatisfied for en0
`,
			wantCount: 1,
			wantTypes: []string{"path_unsatisfied"},
		},
		{
			name: "path satisfied_change should not match path_satisfied",
			input: `2024-01-15 10:32:00.123 Df trustd[554:3333] [com.apple.network:connection] [C5 ready parent-flow (satisfied)] event: path:satisfied_change @0.165s
`,
			wantCount: 0,
			wantTypes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseNetworkEvents(tt.input)

			if len(got) != tt.wantCount {
				t.Fatalf("ParseNetworkEvents() returned %d events, want %d", len(got), tt.wantCount)
			}

			for i, wantType := range tt.wantTypes {
				if got[i].Type != wantType {
					t.Errorf("event[%d].Type = %q, want %q", i, got[i].Type, wantType)
				}
			}
		})
	}
}

func TestParseNetworkEvents_InterfaceExtraction(t *testing.T) {
	input := `2024-01-15 10:30:00.123 Df wifid[234:1111] [com.apple.wifi:manager] WiFi disassociated on en0
`
	events := ParseNetworkEvents(input)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Interface != "en0" {
		t.Errorf("Interface = %q, want %q", events[0].Interface, "en0")
	}
}

func TestParseNetworkEvents_Timestamp(t *testing.T) {
	input := `2024-01-15 10:30:45.123 Df wifid[234:1111] [com.apple.wifi:manager] WiFi disassociated
`
	events := ParseNetworkEvents(input)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Timestamp != "2024-01-15 10:30:45.123" {
		t.Errorf("Timestamp = %q, want %q", events[0].Timestamp, "2024-01-15 10:30:45.123")
	}
}

func TestParseNetworkEvents_VPNEvents(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{
			name:     "VPN connect",
			input:    `2024-01-15 10:30:00.123 Df nesessionmanager[500:1111] [com.apple.network:vpn] VPN connection established on utun0`,
			wantType: "vpn_connect",
		},
		{
			name:     "VPN disconnect",
			input:    `2024-01-15 10:30:00.123 Df nesessionmanager[500:1111] [com.apple.network:vpn] VPN disconnected on utun0`,
			wantType: "vpn_disconnect",
		},
		{
			name:     "tunnel established",
			input:    `2024-01-15 10:30:00.123 Df nesessionmanager[500:1111] [com.apple.network:vpn] tunnel established for utun0`,
			wantType: "vpn_connect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := ParseNetworkEvents(tt.input)
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Type != tt.wantType {
				t.Errorf("Type = %q, want %q", events[0].Type, tt.wantType)
			}
		})
	}
}

func TestNetworkEventJSON(t *testing.T) {
	event := NetworkEvent{
		Timestamp: "2024-01-15 10:30:00.123",
		Type:      "wifi_disconnect",
		Interface: "en0",
		Detail:    "WiFi disassociated from network",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"timestamp", "type", "interface", "detail"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestNetworkEventJSON_OmitEmpty(t *testing.T) {
	event := NetworkEvent{
		Timestamp: "2024-01-15 10:30:00.123",
		Type:      "dns_change",
		Detail:    "DNS config changed",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := parsed["interface"]; ok {
		t.Error("expected 'interface' to be omitted when empty")
	}
}

func TestSummarizeMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short message unchanged",
			input: "WiFi disassociated",
			want:  "WiFi disassociated",
		},
		{
			name:  "strips compact log prefix with connection info",
			input: "Df mDNSResponder[461:dbcd2b] [com.apple.network:connection] [C5666 example.com:443] path:satisfied",
			want:  "[C5666 example.com:443] path:satisfied",
		},
		{
			name:  "strips compact log prefix leaving plain message",
			input: "E  symptomsd[234:abc123] [com.apple.network:wifi] WiFi disconnected",
			want:  "WiFi disconnected",
		},
		{
			name:  "no prefix to strip",
			input: "Simple message",
			want:  "Simple message",
		},
		{
			name:  "long message after prefix strip still truncated",
			input: "Df proc[1:2] [a:b] " + string(make([]byte, 300)),
			want:  string(make([]byte, 197)) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeMessage(tt.input)
			if got != tt.want {
				t.Errorf("summarizeMessage() = %q, want %q", got, tt.want)
			}
			if len(got) > 200 {
				t.Errorf("summarizeMessage() returned %d chars, want <= 200", len(got))
			}
		})
	}
}

func TestEventService_GetEvents(t *testing.T) {
	mock := &MockCmdRunner{
		Output: []byte(`2024-01-15 10:30:00.123 Df wifid[234:1111] [com.apple.wifi:manager] WiFi disassociated on en0
2024-01-15 10:31:00.123 Df wifid[234:2222] [com.apple.wifi:manager] WiFi associated with SSID MyNetwork on en0
`),
	}

	svc := NewEventService(mock)
	events, err := svc.GetEvents(context.Background(), "24h")
	if err != nil {
		t.Fatalf("GetEvents() error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}

func TestEventService_GetEvents_Error(t *testing.T) {
	mock := &MockCmdRunner{
		Err: fmt.Errorf("log command failed"),
	}

	svc := NewEventService(mock)
	_, err := svc.GetEvents(context.Background(), "24h")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeduplicateEvents(t *testing.T) {
	tests := []struct {
		name       string
		events     []NetworkEvent
		window     time.Duration
		wantCount  int
		wantCounts []int
	}{
		{
			name:      "empty input",
			events:    nil,
			window:    30 * time.Second,
			wantCount: 0,
		},
		{
			name: "single event gets count 1",
			events: []NetworkEvent{
				{Timestamp: "2025-01-15 10:30:00.123", Type: "path_satisfied", Detail: "test"},
			},
			window:     30 * time.Second,
			wantCount:  1,
			wantCounts: []int{1},
		},
		{
			name: "5 consecutive same-type within window collapsed",
			events: []NetworkEvent{
				{Timestamp: "2025-01-15 10:30:00.123", Type: "path_satisfied", Detail: "d1"},
				{Timestamp: "2025-01-15 10:30:05.123", Type: "path_satisfied", Detail: "d2"},
				{Timestamp: "2025-01-15 10:30:10.123", Type: "path_satisfied", Detail: "d3"},
				{Timestamp: "2025-01-15 10:30:15.123", Type: "path_satisfied", Detail: "d4"},
				{Timestamp: "2025-01-15 10:30:20.123", Type: "path_satisfied", Detail: "d5"},
			},
			window:     30 * time.Second,
			wantCount:  1,
			wantCounts: []int{5},
		},
		{
			name: "2 different types remain separate",
			events: []NetworkEvent{
				{Timestamp: "2025-01-15 10:30:00.123", Type: "path_satisfied", Detail: "d1"},
				{Timestamp: "2025-01-15 10:30:05.123", Type: "wifi_disconnect", Detail: "d2"},
			},
			window:     30 * time.Second,
			wantCount:  2,
			wantCounts: []int{1, 1},
		},
		{
			name: "same type but outside window produces separate events",
			events: []NetworkEvent{
				{Timestamp: "2025-01-15 10:30:00.123", Type: "path_satisfied", Detail: "d1"},
				{Timestamp: "2025-01-15 10:31:00.123", Type: "path_satisfied", Detail: "d2"},
			},
			window:     30 * time.Second,
			wantCount:  2,
			wantCounts: []int{1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeduplicateEvents(tt.events, tt.window)

			if len(got) != tt.wantCount {
				t.Fatalf("DeduplicateEvents() returned %d events, want %d", len(got), tt.wantCount)
			}

			for i, wantCount := range tt.wantCounts {
				if got[i].Count != wantCount {
					t.Errorf("event[%d].Count = %d, want %d", i, got[i].Count, wantCount)
				}
			}
		})
	}
}
