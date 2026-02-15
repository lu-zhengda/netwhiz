package network

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestParseVPNList(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   []VPNConnection
	}{
		{
			name: "single connected VPN",
			input: `* (Connected)      12345678-1234-1234-1234-123456789012 IPSec              "My Work VPN"
`,
			want: []VPNConnection{
				{Name: "My Work VPN", Type: "IPSec", Status: "Connected"},
			},
		},
		{
			name: "multiple VPNs mixed status",
			input: `* (Connected)      12345678-1234-1234-1234-123456789012 IPSec              "Work VPN"
* (Disconnected)   87654321-4321-4321-4321-210987654321 IKEv2              "Home VPN"
* (Disconnected)   abcdef01-2345-6789-abcd-ef0123456789 L2TP               "Office VPN"
`,
			want: []VPNConnection{
				{Name: "Work VPN", Type: "IPSec", Status: "Connected"},
				{Name: "Home VPN", Type: "IKEv2", Status: "Disconnected"},
				{Name: "Office VPN", Type: "L2TP", Status: "Disconnected"},
			},
		},
		{
			name:  "empty output",
			input: "",
			want:  nil,
		},
		{
			name:  "no VPNs configured",
			input: "\n",
			want:  nil,
		},
		{
			name: "VPN name with special characters",
			input: `* (Disconnected)   12345678-1234-1234-1234-123456789012 IKEv2              "My (Company) VPN - US East"
`,
			want: []VPNConnection{
				{Name: "My (Company) VPN - US East", Type: "IKEv2", Status: "Disconnected"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseVPNList(tt.input)

			if len(got) != len(tt.want) {
				t.Fatalf("ParseVPNList() returned %d items, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("vpn[%d].Name = %q, want %q", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Type != tt.want[i].Type {
					t.Errorf("vpn[%d].Type = %q, want %q", i, got[i].Type, tt.want[i].Type)
				}
				if got[i].Status != tt.want[i].Status {
					t.Errorf("vpn[%d].Status = %q, want %q", i, got[i].Status, tt.want[i].Status)
				}
			}
		})
	}
}

func TestParseVPNStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		vpnName  string
		vpnType  string
		wantStatus   string
		wantServer   string
		wantRemoteIP string
	}{
		{
			name: "connected with details",
			input: `Connected
  ServerAddress : vpn.example.com
  RemoteAddress : 10.0.0.1
  ConnectTime : 1700000000
`,
			vpnName:      "Work VPN",
			vpnType:      "IPSec",
			wantStatus:   "Connected",
			wantServer:   "vpn.example.com",
			wantRemoteIP: "10.0.0.1",
		},
		{
			name: "disconnected",
			input: `Disconnected
  No extended status is available
`,
			vpnName:    "Home VPN",
			vpnType:    "IKEv2",
			wantStatus: "Disconnected",
		},
		{
			name: "connecting",
			input: `Connecting
  ServerAddress : vpn.corp.com
`,
			vpnName:    "Corp VPN",
			vpnType:    "L2TP",
			wantStatus: "Connecting",
			wantServer: "vpn.corp.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseVPNStatus(tt.input, tt.vpnName, tt.vpnType)

			if got.Name != tt.vpnName {
				t.Errorf("Name = %q, want %q", got.Name, tt.vpnName)
			}
			if got.Type != tt.vpnType {
				t.Errorf("Type = %q, want %q", got.Type, tt.vpnType)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", got.Status, tt.wantStatus)
			}
			if got.Server != tt.wantServer {
				t.Errorf("Server = %q, want %q", got.Server, tt.wantServer)
			}
			if got.RemoteIP != tt.wantRemoteIP {
				t.Errorf("RemoteIP = %q, want %q", got.RemoteIP, tt.wantRemoteIP)
			}
		})
	}
}

func TestVPNConnectionJSON(t *testing.T) {
	conn := VPNConnection{
		Name:     "Work VPN",
		Type:     "IKEv2",
		Status:   "Connected",
		Server:   "vpn.example.com",
		Uptime:   "2h 30m 15s",
		RemoteIP: "10.0.0.1",
	}

	data, err := json.Marshal(conn)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"name", "type", "status", "server", "uptime", "remote_ip"}
	for _, key := range expectedKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestVPNConnectionJSON_OmitEmpty(t *testing.T) {
	conn := VPNConnection{
		Name:   "Simple VPN",
		Type:   "IPSec",
		Status: "Disconnected",
	}

	data, err := json.Marshal(conn)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// These should be omitted when empty.
	for _, key := range []string{"server", "uptime", "remote_ip"} {
		if _, ok := parsed[key]; ok {
			t.Errorf("expected key %q to be omitted when empty", key)
		}
	}

	// These should always be present.
	for _, key := range []string{"name", "type", "status"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("expected key %q in JSON output", key)
		}
	}
}

func TestVPNService_ListVPNs(t *testing.T) {
	mock := &MockCmdRunner{
		Output: []byte(`* (Connected)      12345678-1234-1234-1234-123456789012 IPSec              "Work VPN"
* (Disconnected)   87654321-4321-4321-4321-210987654321 IKEv2              "Home VPN"
`),
	}

	svc := NewVPNService(mock)
	vpns, err := svc.ListVPNs(context.Background())
	if err != nil {
		t.Fatalf("ListVPNs() error: %v", err)
	}

	if len(vpns) != 2 {
		t.Fatalf("expected 2 VPNs, got %d", len(vpns))
	}

	if vpns[0].Name != "Work VPN" || vpns[0].Status != "Connected" {
		t.Errorf("unexpected first VPN: %+v", vpns[0])
	}
}

func TestVPNService_ListVPNs_Error(t *testing.T) {
	mock := &MockCmdRunner{
		Err: fmt.Errorf("command failed"),
	}

	svc := NewVPNService(mock)
	_, err := svc.ListVPNs(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "non-numeric",
			input: "not a number",
			want:  "not a number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatUptime(tt.input)
			if got != tt.want {
				t.Errorf("formatUptime(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
