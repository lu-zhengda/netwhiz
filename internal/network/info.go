package network

import (
	"context"
	"fmt"
	"strings"
)

// InfoService gathers system network information.
type InfoService struct {
	runner CmdRunner
}

// NewInfoService creates a new InfoService.
func NewInfoService(runner CmdRunner) *InfoService {
	return &InfoService{runner: runner}
}

// GetNetworkInfo returns the current network interface information.
func (s *InfoService) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	info := &NetworkInfo{}

	// Get active interface.
	iface, err := s.getActiveInterface(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect active interface: %w", err)
	}
	info.Interface = iface

	// Get interface details via networksetup.
	serviceName, err := s.getServiceName(ctx, iface)
	if err != nil {
		serviceName = "Wi-Fi"
	}

	out, err := s.runner.Run(ctx, "networksetup", "-getinfo", serviceName)
	if err == nil {
		s.parseNetworkSetupInfo(string(out), info)
	}

	// Get IP address directly if not set.
	if info.IPAddress == "" {
		out, err = s.runner.Run(ctx, "ipconfig", "getifaddr", iface)
		if err == nil {
			info.IPAddress = strings.TrimSpace(string(out))
		}
	}

	// Get gateway.
	if info.Router == "" {
		out, err = s.runner.Run(ctx, "route", "-n", "get", "default")
		if err == nil {
			info.Router = parseRouteGateway(string(out))
		}
	}

	// Get DNS servers.
	if len(info.DNS) == 0 {
		out, err = s.runner.Run(ctx, "networksetup", "-getdnsservers", serviceName)
		if err == nil {
			info.DNS = parseDNSServers(string(out))
		}
	}

	// Get MAC address.
	out, err = s.runner.Run(ctx, "ifconfig", iface)
	if err == nil {
		info.MACAddress = parseMAC(string(out))
	}

	// Determine status.
	if info.IPAddress != "" {
		info.Status = "Active"
	} else {
		info.Status = "Inactive"
	}

	// Get public IP.
	out, err = s.runner.Run(ctx, "curl", "-s", "--max-time", "3", "https://api.ipify.org")
	if err == nil {
		ip := strings.TrimSpace(string(out))
		if isValidIP(ip) {
			info.PublicIP = ip
		}
	}

	return info, nil
}

// getActiveInterface detects the primary network interface.
func (s *InfoService) getActiveInterface(ctx context.Context) (string, error) {
	out, err := s.runner.Run(ctx, "route", "-n", "get", "default")
	if err != nil {
		return "en0", nil // fallback
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "interface:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "en0", nil
}

// getServiceName maps an interface name to its network service name.
func (s *InfoService) getServiceName(ctx context.Context, iface string) (string, error) {
	out, err := s.runner.Run(ctx, "networksetup", "-listallhardwareports")
	if err != nil {
		return "", fmt.Errorf("failed to list hardware ports: %w", err)
	}

	lines := strings.Split(string(out), "\n")
	var currentService string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Hardware Port:") {
			currentService = strings.TrimPrefix(line, "Hardware Port: ")
		}
		if strings.HasPrefix(line, "Device:") {
			device := strings.TrimPrefix(line, "Device: ")
			device = strings.TrimSpace(device)
			if device == iface {
				return currentService, nil
			}
		}
	}

	return "Wi-Fi", nil
}

// parseNetworkSetupInfo parses the output of `networksetup -getinfo`.
func (s *InfoService) parseNetworkSetupInfo(output string, info *NetworkInfo) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "IP address:") {
			info.IPAddress = strings.TrimSpace(strings.TrimPrefix(line, "IP address:"))
		} else if strings.HasPrefix(line, "Subnet mask:") {
			info.SubnetMask = strings.TrimSpace(strings.TrimPrefix(line, "Subnet mask:"))
		} else if strings.HasPrefix(line, "Router:") {
			info.Router = strings.TrimSpace(strings.TrimPrefix(line, "Router:"))
		}
	}
}

// parseRouteGateway extracts the gateway from `route -n get default` output.
func parseRouteGateway(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gateway:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
		}
	}
	return ""
}

// parseDNSServers parses DNS servers from networksetup output.
func parseDNSServers(output string) []string {
	var servers []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "There aren't any") {
			continue
		}
		if isValidIP(line) {
			servers = append(servers, line)
		}
	}
	return servers
}

// parseMAC extracts the MAC address from ifconfig output.
func parseMAC(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ether ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

// isValidIP performs a basic check for a valid IP address string.
func isValidIP(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// Simple validation: should contain dots and only digits/dots/colons.
	parts := strings.Split(s, ".")
	if len(parts) == 4 {
		for _, p := range parts {
			if p == "" {
				return false
			}
			for _, c := range p {
				if c < '0' || c > '9' {
					return false
				}
			}
		}
		return true
	}
	// IPv6 check (contains colons).
	return strings.Contains(s, ":")
}
