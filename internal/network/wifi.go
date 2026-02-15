package network

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const airportBin = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"

// airportAvailable checks if the legacy airport binary exists.
func airportAvailable() bool {
	_, err := os.Stat(airportBin)
	return err == nil
}

// WiFiService gathers WiFi information.
type WiFiService struct {
	runner CmdRunner
}

// NewWiFiService creates a new WiFiService.
func NewWiFiService(runner CmdRunner) *WiFiService {
	return &WiFiService{runner: runner}
}

// GetWiFiInfo returns current WiFi connection details.
func (s *WiFiService) GetWiFiInfo(ctx context.Context) (*WiFiInfo, error) {
	if airportAvailable() {
		return s.getWiFiInfoAirport(ctx)
	}
	return s.getWiFiInfoProfiler(ctx)
}

// getWiFiInfoAirport uses the legacy airport binary.
func (s *WiFiService) getWiFiInfoAirport(ctx context.Context) (*WiFiInfo, error) {
	out, err := s.runner.Run(ctx, airportBin, "-I")
	if err != nil {
		return nil, fmt.Errorf("failed to get WiFi info: %w", err)
	}

	info := &WiFiInfo{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "SSID":
			info.SSID = val
		case "BSSID":
			info.BSSID = val
		case "channel":
			info.Channel, info.Band = parseChannel(val)
		case "agrCtlRSSI":
			info.RSSI, _ = strconv.Atoi(val)
		case "agrCtlNoise":
			info.Noise, _ = strconv.Atoi(val)
		case "lastTxRate":
			info.TxRate, _ = strconv.Atoi(val)
		case "link auth":
			info.Security = val
		case "country_code":
			info.CountryCode = val
		}
	}

	if info.RSSI != 0 && info.Noise != 0 {
		info.SNR = info.RSSI - info.Noise
	}

	return info, nil
}

// getWiFiInfoProfiler falls back to system_profiler for newer macOS.
func (s *WiFiService) getWiFiInfoProfiler(ctx context.Context) (*WiFiInfo, error) {
	out, err := s.runner.Run(ctx, "system_profiler", "SPAirPortDataType", "-detailLevel", "basic")
	if err != nil {
		return nil, fmt.Errorf("failed to get WiFi info via system_profiler: %w", err)
	}

	info := &WiFiInfo{}
	lines := strings.Split(string(out), "\n")
	inCurrentNetwork := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Current Network Information:") {
			inCurrentNetwork = true
			continue
		}
		if !inCurrentNetwork {
			continue
		}

		// The SSID is the first indented line after "Current Network Information:"
		// e.g. "  MyNetwork:" — it ends with ":" but has no " : " key-value separator.
		if info.SSID == "" && strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, ": ") {
			info.SSID = strings.TrimSuffix(trimmed, ":")
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Channel":
			info.Channel, info.Band = parseChannel(val)
		case "Signal / Noise":
			// Format: "-45 dBm / -90 dBm"
			snParts := strings.Split(val, "/")
			if len(snParts) == 2 {
				sigStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(snParts[0]), "dBm"))
				noiseStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(snParts[1]), "dBm"))
				info.RSSI, _ = strconv.Atoi(sigStr)
				info.Noise, _ = strconv.Atoi(noiseStr)
			}
		case "PHY Mode":
			info.Security = val
		case "Transmit Rate":
			rateStr := strings.TrimSpace(strings.TrimSuffix(val, "Mbps"))
			info.TxRate, _ = strconv.Atoi(rateStr)
		case "Security":
			info.Security = val
		case "Country Code":
			info.CountryCode = val
		}
	}

	if info.RSSI != 0 && info.Noise != 0 {
		info.SNR = info.RSSI - info.Noise
	}

	// system_profiler redacts SSID on newer macOS; supplement with networksetup.
	if info.SSID == "" || info.SSID == "<redacted>" {
		if ssid, err := s.getSSIDFromNetworksetup(ctx); err == nil && ssid != "" {
			info.SSID = ssid
		}
	}

	return info, nil
}

// getSSIDFromNetworksetup retrieves the current WiFi SSID via networksetup.
func (s *WiFiService) getSSIDFromNetworksetup(ctx context.Context) (string, error) {
	out, err := s.runner.Run(ctx, "networksetup", "-getairportnetwork", "en0")
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(string(out))
	const prefix = "Current Wi-Fi Network: "
	if strings.HasPrefix(line, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(line, prefix)), nil
	}
	return "", fmt.Errorf("unexpected networksetup output: %s", line)
}

// ScanNetworks scans for nearby WiFi networks.
func (s *WiFiService) ScanNetworks(ctx context.Context) ([]WiFiNetwork, error) {
	if !airportAvailable() {
		return nil, fmt.Errorf("WiFi scan requires the airport utility (removed in macOS 15+); use 'networksetup -listpreferredwirelessnetworks en0' instead")
	}
	out, err := s.runner.Run(ctx, airportBin, "-s")
	if err != nil {
		return nil, fmt.Errorf("failed to scan WiFi networks: %w", err)
	}

	return parseAirportScan(string(out)), nil
}

// parseChannel parses a channel string like "149", "149,+1", or "52 (5GHz, 80MHz)" into channel number and band.
func parseChannel(s string) (int, string) {
	// Handle formats: "149", "149,+1", "52 (5GHz, 80MHz)".
	// Split on space first to handle system_profiler format, then on comma for airport format.
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0, "Unknown"
	}
	parts := strings.Split(fields[0], ",")
	ch, _ := strconv.Atoi(strings.TrimSpace(parts[0]))

	var band string
	switch {
	case ch >= 1 && ch <= 14:
		band = "2.4GHz"
	case ch >= 36 && ch <= 177:
		band = "5GHz"
	case ch >= 1 && ch <= 233: // Wi-Fi 6E uses channels > 177 in 6GHz
		band = "6GHz"
	default:
		band = "Unknown"
	}

	// Correct: channels 36-177 are 5GHz, channels above 177 are 6GHz.
	if ch > 177 {
		band = "6GHz"
	}

	return ch, band
}

// parseAirportScan parses the output of `airport -s`.
func parseAirportScan(output string) []WiFiNetwork {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return nil
	}

	var networks []WiFiNetwork
	// Skip header line.
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Airport -s output is fixed-width. Parse carefully.
		// Format: SSID BSSID RSSI CHANNEL HT CC SECURITY
		// The SSID can contain spaces, so we parse from the right.
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		// Find BSSID (contains colons like aa:bb:cc:dd:ee:ff).
		bssidIdx := -1
		for i, f := range fields {
			if len(f) == 17 && strings.Count(f, ":") == 5 {
				bssidIdx = i
				break
			}
		}
		if bssidIdx < 0 {
			continue
		}

		ssid := strings.Join(fields[:bssidIdx], " ")
		bssid := fields[bssidIdx]

		rssi := 0
		channel := 0
		security := ""

		if bssidIdx+1 < len(fields) {
			rssi, _ = strconv.Atoi(fields[bssidIdx+1])
		}
		if bssidIdx+2 < len(fields) {
			ch := strings.Split(fields[bssidIdx+2], ",")
			channel, _ = strconv.Atoi(ch[0])
		}
		if bssidIdx+5 < len(fields) {
			security = strings.Join(fields[bssidIdx+5:], " ")
		}

		networks = append(networks, WiFiNetwork{
			SSID:     ssid,
			BSSID:    bssid,
			RSSI:     rssi,
			Channel:  channel,
			Security: security,
		})
	}

	return networks
}

// SignalQuality returns a human-readable signal quality description for an RSSI value.
func SignalQuality(rssi int) string {
	switch {
	case rssi >= -50:
		return "Excellent"
	case rssi >= -60:
		return "Good"
	case rssi >= -70:
		return "Fair"
	case rssi >= -80:
		return "Weak"
	default:
		return "Poor"
	}
}

// SignalBar returns a visual signal strength bar for an RSSI value.
func SignalBar(rssi int) string {
	switch {
	case rssi >= -50:
		return "\u2588\u2588\u2588\u2588\u2591" // "Excellent"
	case rssi >= -60:
		return "\u2588\u2588\u2588\u2591\u2591" // "Good"
	case rssi >= -70:
		return "\u2588\u2588\u2591\u2591\u2591" // "Fair"
	case rssi >= -80:
		return "\u2588\u2591\u2591\u2591\u2591" // "Weak"
	default:
		return "\u2591\u2591\u2591\u2591\u2591" // "Poor"
	}
}
