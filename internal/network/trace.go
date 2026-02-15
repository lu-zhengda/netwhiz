package network

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TraceService wraps the system traceroute command.
type TraceService struct {
	runner CmdRunner
}

// NewTraceService creates a new TraceService.
func NewTraceService(runner CmdRunner) *TraceService {
	return &TraceService{runner: runner}
}

// Traceroute runs a traceroute to the given host.
func (s *TraceService) Traceroute(ctx context.Context, host string) ([]TraceHop, error) {
	out, err := s.runner.Run(ctx, "traceroute", host)
	if err != nil {
		// traceroute may return non-zero but still produce useful output.
		if out == nil {
			return nil, fmt.Errorf("failed to run traceroute: %w", err)
		}
	}

	return parseTraceroute(string(out)), nil
}

// parseTraceroute parses the output of the system traceroute command.
func parseTraceroute(output string) []TraceHop {
	var hops []TraceHop

	// Traceroute output format:
	// 1  router.local (192.168.1.1)  1.234 ms  1.456 ms  1.789 ms
	// 2  * * *
	// 3  10.0.0.1 (10.0.0.1)  5.678 ms  5.890 ms  6.012 ms

	lines := strings.Split(output, "\n")
	rttRe := regexp.MustCompile(`([\d.]+)\s+ms`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line starts with a hop number.
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		hopNum, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		hop := TraceHop{Hop: hopNum}

		// Check for all-stars (lost).
		if strings.Contains(line, "* * *") {
			hop.Lost = true
			hops = append(hops, hop)
			continue
		}

		// Extract IP and hostname.
		ipRe := regexp.MustCompile(`\(([\d.]+)\)`)
		ipMatches := ipRe.FindStringSubmatch(line)
		if len(ipMatches) >= 2 {
			hop.IP = ipMatches[1]
		}

		// Hostname is the second field if it's not an IP in parens.
		if len(fields) >= 2 {
			hostname := fields[1]
			if hostname != "*" && !strings.HasPrefix(hostname, "(") {
				hop.Hostname = hostname
			}
		}

		// If no hostname but we have IP, use IP as hostname.
		if hop.Hostname == "" && hop.IP != "" {
			hop.Hostname = hop.IP
		}

		// Extract RTTs.
		rttMatches := rttRe.FindAllStringSubmatch(line, -1)
		for _, match := range rttMatches {
			ms, err := strconv.ParseFloat(match[1], 64)
			if err == nil {
				hop.RTTs = append(hop.RTTs, time.Duration(ms*float64(time.Millisecond)))
			}
		}

		hops = append(hops, hop)
	}

	return hops
}
