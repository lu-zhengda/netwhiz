package network

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PingService wraps the system ping command.
type PingService struct {
	runner CmdRunner
}

// NewPingService creates a new PingService.
func NewPingService(runner CmdRunner) *PingService {
	return &PingService{runner: runner}
}

// Ping runs a ping command against the given host with the specified count.
func (s *PingService) Ping(ctx context.Context, host string, count int) (*PingStats, error) {
	if count <= 0 {
		count = 5
	}

	out, err := s.runner.Run(ctx, "ping", "-c", strconv.Itoa(count), host)
	if err != nil {
		// ping may return non-zero if packets are lost, but output is still useful.
		if out == nil {
			return nil, fmt.Errorf("failed to run ping: %w", err)
		}
	}

	return parsePingOutput(string(out), host)
}

// parsePingOutput parses the output of the system ping command.
func parsePingOutput(output string, host string) (*PingStats, error) {
	stats := &PingStats{Host: host}

	// Parse individual ping lines.
	// Format: 64 bytes from 142.250.80.46: icmp_seq=0 ttl=117 time=12.345 ms
	pingLineRe := regexp.MustCompile(`icmp_seq=(\d+)\s+ttl=(\d+)\s+time=([\d.]+)\s*ms`)
	for _, line := range strings.Split(output, "\n") {
		matches := pingLineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		seq, _ := strconv.Atoi(matches[1])
		ttl, _ := strconv.Atoi(matches[2])
		timeMs, _ := strconv.ParseFloat(matches[3], 64)

		stats.Results = append(stats.Results, PingResult{
			Host:      host,
			Seq:       seq,
			TTL:       ttl,
			Time:      time.Duration(timeMs * float64(time.Millisecond)),
			Timestamp: time.Now(),
		})
	}

	// Parse statistics line.
	// Format: 5 packets transmitted, 5 packets received, 0.0% packet loss
	statsRe := regexp.MustCompile(`(\d+) packets transmitted, (\d+) (?:packets )?received, ([\d.]+)% packet loss`)
	for _, line := range strings.Split(output, "\n") {
		matches := statsRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		stats.Sent, _ = strconv.Atoi(matches[1])
		stats.Received, _ = strconv.Atoi(matches[2])
		stats.LossPercent, _ = strconv.ParseFloat(matches[3], 64)
		stats.Lost = stats.Sent - stats.Received
	}

	// Parse round-trip stats.
	// Format: round-trip min/avg/max/stddev = 11.8/18.9/45.2/13.4 ms
	rttRe := regexp.MustCompile(`round-trip min/avg/max/stddev = ([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+) ms`)
	for _, line := range strings.Split(output, "\n") {
		matches := rttRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		minMs, _ := strconv.ParseFloat(matches[1], 64)
		avgMs, _ := strconv.ParseFloat(matches[2], 64)
		maxMs, _ := strconv.ParseFloat(matches[3], 64)
		stdMs, _ := strconv.ParseFloat(matches[4], 64)

		stats.MinRTT = time.Duration(minMs * float64(time.Millisecond))
		stats.AvgRTT = time.Duration(avgMs * float64(time.Millisecond))
		stats.MaxRTT = time.Duration(maxMs * float64(time.Millisecond))
		stats.StdDevRTT = time.Duration(stdMs * float64(time.Millisecond))
	}

	// If no RTT stats were parsed from the summary, calculate from results.
	if stats.MinRTT == 0 && len(stats.Results) > 0 {
		stats.MinRTT, stats.MaxRTT, stats.AvgRTT, stats.StdDevRTT = calculateRTTStats(stats.Results)
	}

	return stats, nil
}

// calculateRTTStats computes min/max/avg/stddev from ping results.
func calculateRTTStats(results []PingResult) (min, max, avg, stddev time.Duration) {
	if len(results) == 0 {
		return
	}

	min = results[0].Time
	max = results[0].Time
	var total float64

	for _, r := range results {
		if r.Time < min {
			min = r.Time
		}
		if r.Time > max {
			max = r.Time
		}
		total += float64(r.Time)
	}

	avgFloat := total / float64(len(results))
	avg = time.Duration(avgFloat)

	if len(results) > 1 {
		var sumSq float64
		for _, r := range results {
			diff := float64(r.Time) - avgFloat
			sumSq += diff * diff
		}
		stddev = time.Duration(math.Sqrt(sumSq / float64(len(results))))
	}

	return
}

// PingBar renders a visual bar for a ping time relative to the max observed time.
func PingBar(rtt, maxRTT time.Duration, width int) string {
	if maxRTT == 0 || width <= 0 {
		return ""
	}

	ratio := float64(rtt) / float64(maxRTT)
	filled := int(ratio * float64(width))
	if filled == 0 && rtt > 0 {
		filled = 1
	}

	blocks := []rune("\u2588") // full block
	bar := strings.Repeat(string(blocks), filled) + strings.Repeat(" ", width-filled)
	return "\u258f" + bar + "\u258f" // left bar + content + right bar
}
