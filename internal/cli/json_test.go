package cli

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONAction_Serialization(t *testing.T) {
	tests := []struct {
		name   string
		action jsonAction
		want   map[string]interface{}
	}{
		{
			name: "dns set",
			action: jsonAction{
				OK:      true,
				Action:  "dns_set",
				Target:  "cloudflare",
				Message: "DNS set to cloudflare (1.1.1.1, 1.0.0.1)",
			},
			want: map[string]interface{}{
				"ok":      true,
				"action":  "dns_set",
				"target":  "cloudflare",
				"message": "DNS set to cloudflare (1.1.1.1, 1.0.0.1)",
			},
		},
		{
			name: "dns flush",
			action: jsonAction{
				OK:      true,
				Action:  "dns_flush",
				Message: "DNS cache flushed successfully",
			},
			want: map[string]interface{}{
				"ok":      true,
				"action":  "dns_flush",
				"message": "DNS cache flushed successfully",
			},
		},
		{
			name: "vpn connect",
			action: jsonAction{
				OK:      true,
				Action:  "vpn_connect",
				Target:  "My VPN",
				Message: `VPN "My VPN" status: Connected`,
			},
			want: map[string]interface{}{
				"ok":      true,
				"action":  "vpn_connect",
				"target":  "My VPN",
				"message": `VPN "My VPN" status: Connected`,
			},
		},
		{
			name: "vpn disconnect",
			action: jsonAction{
				OK:      true,
				Action:  "vpn_disconnect",
				Target:  "My VPN",
				Message: `Disconnected "My VPN"`,
			},
			want: map[string]interface{}{
				"ok":      true,
				"action":  "vpn_disconnect",
				"target":  "My VPN",
				"message": `Disconnected "My VPN"`,
			},
		},
		{
			name: "vpn disconnect all",
			action: jsonAction{
				OK:      true,
				Action:  "vpn_disconnect_all",
				Message: "Disconnected 2 VPN(s)",
			},
			want: map[string]interface{}{
				"ok":      true,
				"action":  "vpn_disconnect_all",
				"message": "Disconnected 2 VPN(s)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := fprintJSON(&buf, tt.action); err != nil {
				t.Fatalf("fprintJSON returned error: %v", err)
			}

			var got map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
			}

			for key, wantVal := range tt.want {
				gotVal, ok := got[key]
				if !ok {
					t.Errorf("expected key %q in JSON output", key)
					continue
				}
				// json.Unmarshal decodes booleans as bool and strings as string.
				switch wv := wantVal.(type) {
				case bool:
					if gv, ok := gotVal.(bool); !ok || gv != wv {
						t.Errorf("key %q: got %v, want %v", key, gotVal, wantVal)
					}
				case string:
					if gv, ok := gotVal.(string); !ok || gv != wv {
						t.Errorf("key %q: got %v, want %v", key, gotVal, wantVal)
					}
				}
			}
		})
	}
}

func TestJSONAction_OmitEmptyTarget(t *testing.T) {
	action := jsonAction{
		OK:      true,
		Action:  "dns_flush",
		Message: "DNS cache flushed successfully",
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, action); err != nil {
		t.Fatalf("fprintJSON returned error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if _, ok := got["target"]; ok {
		t.Error("expected target to be omitted when empty")
	}
}

func TestJSONAction_OmitEmptyMessage(t *testing.T) {
	action := jsonAction{
		OK:     true,
		Action: "test",
		Target: "something",
	}

	var buf bytes.Buffer
	if err := fprintJSON(&buf, action); err != nil {
		t.Fatalf("fprintJSON returned error: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if _, ok := got["message"]; ok {
		t.Error("expected message to be omitted when empty")
	}
}
