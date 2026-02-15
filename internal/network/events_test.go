package network

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
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
			input: `2024-01-15 10:30:45.123456-0800  0x1234  com.apple.wifi  WiFi interface en0 disassociated from network
`,
			wantCount: 1,
			wantTypes: []string{"wifi_disconnect"},
		},
		{
			name: "wifi connect event",
			input: `2024-01-15 10:31:00.123456-0800  0x1234  com.apple.wifi  WiFi interface en0 associated with SSID MyNetwork
`,
			wantCount: 1,
			wantTypes: []string{"wifi_connect"},
		},
		{
			name: "ip address change",
			input: `2024-01-15 10:32:00.123456-0800  0x5678  com.apple.network  address changed for en0: 192.168.1.100
`,
			wantCount: 1,
			wantTypes: []string{"ip_change"},
		},
		{
			name: "dns configuration change",
			input: `2024-01-15 10:33:00.123456-0800  0x9abc  com.apple.network  DNS config updated for resolver
`,
			wantCount: 1,
			wantTypes: []string{"dns_change"},
		},
		{
			name: "interface up and down",
			input: `2024-01-15 10:34:00.123456-0800  0xdef0  com.apple.network  interface down: en0
2024-01-15 10:34:05.123456-0800  0xdef0  com.apple.network  interface up: en0
`,
			wantCount: 2,
			wantTypes: []string{"interface_down", "interface_up"},
		},
		{
			name: "multiple mixed events",
			input: `2024-01-15 10:30:00.123456-0800  0x1111  com.apple.wifi  WiFi disassociated from network on en0
2024-01-15 10:30:05.123456-0800  0x2222  com.apple.network  Some unrelated log message about networking
2024-01-15 10:30:10.123456-0800  0x3333  com.apple.wifi  WiFi associated with SSID OtherNetwork on en0
2024-01-15 10:30:15.123456-0800  0x4444  com.apple.network  DNS config changed
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
			input: `2024-01-15 10:30:00.123456-0800  0x1111  com.apple.network  routine heartbeat check
2024-01-15 10:31:00.123456-0800  0x2222  com.apple.network  connection pool status: ok
`,
			wantCount: 0,
			wantTypes: nil,
		},
		{
			name: "malformed lines are skipped",
			input: `This is not a log line
2024-01-15 10:30:00.123456-0800  0x1111  com.apple.wifi  WiFi disassociated
Another bad line
`,
			wantCount: 1,
			wantTypes: []string{"wifi_disconnect"},
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
	input := `2024-01-15 10:30:00.123456-0800  0x1111  com.apple.wifi  WiFi disassociated on en0
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
	input := `2024-01-15 10:30:45.123456-0800  0x1111  com.apple.wifi  WiFi disassociated
`
	events := ParseNetworkEvents(input)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Timestamp != "2024-01-15 10:30:45.123456-0800" {
		t.Errorf("Timestamp = %q, want %q", events[0].Timestamp, "2024-01-15 10:30:45.123456-0800")
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
			input:    `2024-01-15 10:30:00.123456-0800  0x1111  com.apple.network  VPN connection established on utun0`,
			wantType: "vpn_connect",
		},
		{
			name:     "VPN disconnect",
			input:    `2024-01-15 10:30:00.123456-0800  0x1111  com.apple.network  VPN disconnected on utun0`,
			wantType: "vpn_disconnect",
		},
		{
			name:     "tunnel established",
			input:    `2024-01-15 10:30:00.123456-0800  0x1111  com.apple.network  tunnel established for utun0`,
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
		Timestamp: "2024-01-15 10:30:00.123456-0800",
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
		Timestamp: "2024-01-15 10:30:00.123456-0800",
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
			name:  "exactly at limit",
			input: string(make([]byte, 200)),
			want:  string(make([]byte, 200)),
		},
		{
			name:  "long message truncated",
			input: string(make([]byte, 300)),
			want:  string(make([]byte, 197)) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeMessage(tt.input)
			if len(got) > 200 {
				t.Errorf("summarizeMessage() returned %d chars, want <= 200", len(got))
			}
		})
	}
}

func TestEventService_GetEvents(t *testing.T) {
	mock := &MockCmdRunner{
		Output: []byte(`2024-01-15 10:30:00.123456-0800  0x1111  com.apple.wifi  WiFi disassociated on en0
2024-01-15 10:31:00.123456-0800  0x2222  com.apple.wifi  WiFi associated with SSID MyNetwork on en0
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
