package network

import (
	"context"
	"fmt"
	"strings"
)

// DNSService manages DNS configuration.
type DNSService struct {
	runner CmdRunner
}

// NewDNSService creates a new DNSService.
func NewDNSService(runner CmdRunner) *DNSService {
	return &DNSService{runner: runner}
}

// GetDNSServers returns the currently configured DNS servers for the active network service.
func (s *DNSService) GetDNSServers(ctx context.Context) ([]string, string, error) {
	serviceName, err := s.getActiveServiceName(ctx)
	if err != nil {
		serviceName = "Wi-Fi"
	}

	out, err := s.runner.Run(ctx, "networksetup", "-getdnsservers", serviceName)
	if err != nil {
		return nil, serviceName, fmt.Errorf("failed to get DNS servers: %w", err)
	}

	servers := parseDNSServers(string(out))
	return servers, serviceName, nil
}

// SetDNS sets DNS servers for the active network service.
// server can be a preset name (e.g., "cloudflare", "google") or a raw IP.
func (s *DNSService) SetDNS(ctx context.Context, server string) error {
	serviceName, err := s.getActiveServiceName(ctx)
	if err != nil {
		serviceName = "Wi-Fi"
	}

	// Check if it's a preset name.
	var servers []string
	if preset := ResolveDNSPreset(server); preset != nil {
		servers = preset
	} else {
		servers = strings.Split(server, ",")
		for i := range servers {
			servers[i] = strings.TrimSpace(servers[i])
		}
	}

	args := []string{"-setdnsservers", serviceName}
	args = append(args, servers...)

	_, err = s.runner.Run(ctx, "networksetup", args...)
	if err != nil {
		return fmt.Errorf("failed to set DNS servers (may need sudo): %w", err)
	}

	return nil
}

// FlushDNS flushes the macOS DNS cache.
func (s *DNSService) FlushDNS(ctx context.Context) error {
	_, err := s.runner.Run(ctx, "dscacheutil", "-flushcache")
	if err != nil {
		return fmt.Errorf("failed to flush DNS cache: %w", err)
	}

	// Also try to restart mDNSResponder (requires sudo, may fail).
	_, _ = s.runner.Run(ctx, "sudo", "killall", "-HUP", "mDNSResponder")

	return nil
}

// GetDNSConfig returns detailed DNS configuration from scutil.
func (s *DNSService) GetDNSConfig(ctx context.Context) (string, error) {
	out, err := s.runner.Run(ctx, "scutil", "--dns")
	if err != nil {
		return "", fmt.Errorf("failed to get DNS config: %w", err)
	}
	return string(out), nil
}

// getActiveServiceName returns the network service name for the primary interface.
func (s *DNSService) getActiveServiceName(ctx context.Context) (string, error) {
	// Detect active interface first.
	out, err := s.runner.Run(ctx, "route", "-n", "get", "default")
	if err != nil {
		return "Wi-Fi", nil
	}

	var iface string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "interface:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				iface = parts[1]
			}
		}
	}

	if iface == "" {
		return "Wi-Fi", nil
	}

	// Map interface to service name.
	out, err = s.runner.Run(ctx, "networksetup", "-listallhardwareports")
	if err != nil {
		return "Wi-Fi", nil
	}

	lines := strings.Split(string(out), "\n")
	var currentService string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Hardware Port:") {
			currentService = strings.TrimPrefix(line, "Hardware Port: ")
		}
		if strings.HasPrefix(line, "Device:") {
			device := strings.TrimSpace(strings.TrimPrefix(line, "Device: "))
			if device == iface {
				return currentService, nil
			}
		}
	}

	return "Wi-Fi", nil
}

// ResolveDNSPreset returns the servers for a preset name, or nil if not found.
func ResolveDNSPreset(name string) []string {
	lower := strings.ToLower(name)
	for _, p := range DNSPresets {
		if strings.ToLower(p.Name) == lower {
			return p.Servers
		}
	}
	return nil
}

// ListPresets returns all available DNS preset names.
func ListPresets() []string {
	presets := make([]string, 0, len(DNSPresets))
	for _, p := range DNSPresets {
		presets = append(presets, p.Name)
	}
	return presets
}
