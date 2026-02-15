package network

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// VPNConnection holds information about a configured VPN connection.
type VPNConnection struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Server   string `json:"server,omitempty"`
	Uptime   string `json:"uptime,omitempty"`
	RemoteIP string `json:"remote_ip,omitempty"`
}

// VPNService manages VPN connections.
type VPNService struct {
	runner CmdRunner
}

// NewVPNService creates a new VPNService.
func NewVPNService(runner CmdRunner) *VPNService {
	return &VPNService{runner: runner}
}

// ListVPNs returns all configured VPN connections with their status.
func (s *VPNService) ListVPNs(ctx context.Context) ([]VPNConnection, error) {
	out, err := s.runner.Run(ctx, "scutil", "--nc", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list VPN connections: %w", err)
	}
	return ParseVPNList(string(out)), nil
}

// Connect starts a VPN connection by name.
func (s *VPNService) Connect(ctx context.Context, name string) error {
	_, err := s.runner.Run(ctx, "scutil", "--nc", "start", name)
	if err != nil {
		return fmt.Errorf("failed to start VPN %q: %w", name, err)
	}

	// Wait briefly for connection to establish.
	time.Sleep(2 * time.Second)
	return nil
}

// Disconnect stops a VPN connection by name.
func (s *VPNService) Disconnect(ctx context.Context, name string) error {
	_, err := s.runner.Run(ctx, "scutil", "--nc", "stop", name)
	if err != nil {
		return fmt.Errorf("failed to stop VPN %q: %w", name, err)
	}
	return nil
}

// DisconnectAll stops all active VPN connections.
func (s *VPNService) DisconnectAll(ctx context.Context) ([]string, error) {
	vpns, err := s.ListVPNs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list VPNs: %w", err)
	}

	var disconnected []string
	for _, vpn := range vpns {
		if vpn.Status == "Connected" {
			if err := s.Disconnect(ctx, vpn.Name); err != nil {
				return disconnected, fmt.Errorf("failed to disconnect %q: %w", vpn.Name, err)
			}
			disconnected = append(disconnected, vpn.Name)
		}
	}
	return disconnected, nil
}

// GetVPNStatus returns detailed status of a specific VPN connection.
func (s *VPNService) GetVPNStatus(ctx context.Context, name string) (*VPNConnection, error) {
	out, err := s.runner.Run(ctx, "scutil", "--nc", "status", name)
	if err != nil {
		return nil, fmt.Errorf("failed to get VPN status for %q: %w", name, err)
	}

	// Get the VPN list to find the type.
	vpns, _ := s.ListVPNs(ctx)
	var vpnType string
	for _, v := range vpns {
		if v.Name == name {
			vpnType = v.Type
			break
		}
	}

	conn := ParseVPNStatus(string(out), name, vpnType)
	return conn, nil
}

// GetActiveVPNStatus returns detailed status of the first connected VPN.
func (s *VPNService) GetActiveVPNStatus(ctx context.Context) (*VPNConnection, error) {
	vpns, err := s.ListVPNs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list VPNs: %w", err)
	}

	for _, vpn := range vpns {
		if vpn.Status == "Connected" {
			return s.GetVPNStatus(ctx, vpn.Name)
		}
	}

	return nil, fmt.Errorf("no active VPN connection found")
}

// scutilLineRegexp matches lines from `scutil --nc list`.
// Example: * (Connected)      12345678-1234-1234-1234-123456789012 IPSec              "My VPN"
var scutilLineRegexp = regexp.MustCompile(`^\*\s+\((\w+)\)\s+\S+\s+(\S+)\s+"(.+)"`)

// ParseVPNList parses the output of `scutil --nc list`.
func ParseVPNList(output string) []VPNConnection {
	var vpns []VPNConnection

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := scutilLineRegexp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		vpns = append(vpns, VPNConnection{
			Status: matches[1],
			Type:   matches[2],
			Name:   matches[3],
		})
	}

	return vpns
}

// ParseVPNStatus parses the output of `scutil --nc status <name>`.
func ParseVPNStatus(output, name, vpnType string) *VPNConnection {
	conn := &VPNConnection{
		Name: name,
		Type: vpnType,
	}

	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		conn.Status = strings.TrimSpace(lines[0])
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "ServerAddress":
			conn.Server = val
		case "RemoteAddress":
			conn.RemoteIP = val
		case "ConnectTime":
			conn.Uptime = formatUptime(val)
		}
	}

	return conn
}

// formatUptime converts a timestamp or duration value into a human-readable uptime string.
func formatUptime(val string) string {
	// scutil --nc status ConnectTime is a Unix timestamp (seconds since epoch)
	// or an integer representing elapsed seconds.
	val = strings.TrimSpace(val)
	if val == "" {
		return ""
	}

	// Try to parse as an integer (Unix timestamp).
	var ts int64
	n, err := fmt.Sscanf(val, "%d", &ts)
	if err != nil || n != 1 {
		return val
	}

	connectTime := time.Unix(ts, 0)
	elapsed := time.Since(connectTime)
	if elapsed < 0 {
		return val
	}

	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
