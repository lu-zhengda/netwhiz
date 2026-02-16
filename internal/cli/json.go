package cli

// ---------------------------------------------------------------------------
// Action result types for JSON output
// ---------------------------------------------------------------------------

// jsonAction is the standard response for mutating commands
// (dns set, dns flush, vpn connect, vpn disconnect).
type jsonAction struct {
	OK      bool   `json:"ok"`
	Action  string `json:"action"`
	Target  string `json:"target,omitempty"`
	Message string `json:"message,omitempty"`
}
