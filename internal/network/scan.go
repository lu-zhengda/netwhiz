package network

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
)

// ScanService performs ARP network scanning.
type ScanService struct {
	runner CmdRunner
}

// NewScanService creates a new ScanService.
func NewScanService(runner CmdRunner) *ScanService {
	return &ScanService{runner: runner}
}

// ARPScan returns all ARP table entries.
func (s *ScanService) ARPScan(ctx context.Context) ([]ARPEntry, error) {
	out, err := s.runner.Run(ctx, "arp", "-a")
	if err != nil {
		return nil, fmt.Errorf("failed to run ARP scan: %w", err)
	}

	return parseARPOutput(string(out)), nil
}

// parseARPOutput parses the output of `arp -a`.
func parseARPOutput(output string) []ARPEntry {
	var entries []ARPEntry

	// ARP output format:
	// ? (192.168.1.1) at aa:bb:cc:dd:ee:ff on en0 ifscope [ethernet]
	// hostname (192.168.1.100) at 11:22:33:44:55:66 on en0 ifscope [ethernet]
	lineRe := regexp.MustCompile(`^(\S+)\s+\(([\d.]+)\)\s+at\s+(\S+)\s+on\s+(\S+)`)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		hostname := matches[1]
		ip := matches[2]
		mac := matches[3]
		iface := matches[4]

		// Skip incomplete entries.
		if mac == "(incomplete)" {
			continue
		}

		// If hostname is "?", try reverse DNS.
		if hostname == "?" {
			hostname = reverseLookup(ip)
		}

		entries = append(entries, ARPEntry{
			IP:        ip,
			MAC:       mac,
			Interface: iface,
			Hostname:  hostname,
		})
	}

	return entries
}

// reverseLookup performs a reverse DNS lookup for an IP address.
func reverseLookup(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	// Remove trailing dot from DNS name.
	return strings.TrimSuffix(names[0], ".")
}
