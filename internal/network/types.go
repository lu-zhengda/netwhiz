package network

import "time"

// NetworkInfo holds system network interface information.
type NetworkInfo struct {
	Interface  string   `json:"interface"`
	IPAddress  string   `json:"ip_address"`
	SubnetMask string   `json:"subnet_mask"`
	Router     string   `json:"router"`
	DNS        []string `json:"dns"`
	MACAddress string   `json:"mac_address"`
	MediaSpeed string   `json:"media_speed,omitempty"`
	Status     string   `json:"status"` // "active", "inactive"
	PublicIP   string   `json:"public_ip,omitempty"`
}

// WiFiInfo holds current WiFi connection details.
type WiFiInfo struct {
	SSID        string `json:"ssid"`
	BSSID       string `json:"bssid"`
	Channel     int    `json:"channel"`
	Band        string `json:"band"` // "2.4GHz", "5GHz", "6GHz"
	RSSI        int    `json:"rssi"`  // signal strength in dBm
	Noise       int    `json:"noise"` // noise in dBm
	SNR         int    `json:"snr"`   // signal-to-noise ratio
	TxRate      int    `json:"tx_rate"` // Mbps
	Security    string `json:"security"` // "WPA2", "WPA3", etc.
	CountryCode string `json:"country_code"`
}

// WiFiNetwork represents a nearby WiFi network from scanning.
type WiFiNetwork struct {
	SSID     string `json:"ssid"`
	BSSID    string `json:"bssid"`
	RSSI     int    `json:"rssi"`
	Channel  int    `json:"channel"`
	Security string `json:"security"`
}

// PingResult holds a single ping response.
type PingResult struct {
	Host      string        `json:"host"`
	Seq       int           `json:"seq"`
	TTL       int           `json:"ttl"`
	Time      time.Duration `json:"time_ns"`
	TimeMs    float64       `json:"time_ms"`
	Timestamp time.Time     `json:"timestamp"`
}

// PingStats holds aggregate ping statistics.
type PingStats struct {
	Host        string        `json:"host"`
	Sent        int           `json:"sent"`
	Received    int           `json:"received"`
	Lost        int           `json:"lost"`
	LossPercent float64       `json:"loss_percent"`
	MinRTT      time.Duration `json:"min_rtt_ns"`
	MinRTTMs    float64       `json:"min_rtt_ms"`
	MaxRTT      time.Duration `json:"max_rtt_ns"`
	MaxRTTMs    float64       `json:"max_rtt_ms"`
	AvgRTT      time.Duration `json:"avg_rtt_ns"`
	AvgRTTMs    float64       `json:"avg_rtt_ms"`
	StdDevRTT   time.Duration `json:"stddev_rtt_ns"`
	StdDevRTTMs float64       `json:"stddev_rtt_ms"`
	Results     []PingResult  `json:"results"`
}

// TraceHop holds a single traceroute hop.
type TraceHop struct {
	Hop      int             `json:"hop"`
	IP       string          `json:"ip,omitempty"`
	Hostname string          `json:"hostname,omitempty"`
	RTTs     []time.Duration `json:"rtts_ns,omitempty"` // up to 3 probes
	RTTsMs   []float64       `json:"rtts_ms,omitempty"`
	Lost     bool            `json:"lost"`
}

// SpeedResult holds speed test results.
type SpeedResult struct {
	DownloadMbps float64       `json:"download_mbps"`
	UploadMbps   float64       `json:"upload_mbps"`
	Latency      time.Duration `json:"latency_ns"`
	LatencyMs    float64       `json:"latency_ms"`
	Server       string        `json:"server"`
	Timestamp    time.Time     `json:"timestamp"`
}

// ARPEntry holds a single ARP table entry.
type ARPEntry struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Interface string `json:"interface"`
	Hostname  string `json:"hostname,omitempty"` // reverse DNS if available
}

// DNSConfig holds DNS configuration info for JSON output.
type DNSConfig struct {
	ServiceName string      `json:"service_name"`
	Servers     []string    `json:"servers"`
	Preset      string      `json:"preset,omitempty"`
	Presets     []DNSPreset `json:"presets"`
}

// DNSPreset holds a friendly name and DNS server addresses.
type DNSPreset struct {
	Name    string   `json:"name"`
	Servers []string `json:"servers"`
}

// DNSPresets provides available DNS presets in stable display order.
var DNSPresets = []DNSPreset{
	{Name: "cloudflare", Servers: []string{"1.1.1.1", "1.0.0.1"}},
	{Name: "google", Servers: []string{"8.8.8.8", "8.8.4.4"}},
	{Name: "quad9", Servers: []string{"9.9.9.9", "149.112.112.112"}},
	{Name: "opendns", Servers: []string{"208.67.222.222", "208.67.220.220"}},
}

// PingTarget is a preconfigured ping endpoint.
type PingTarget struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

// PingTargets provides preconfigured endpoints for ping.
var PingTargets = []PingTarget{
	{Name: "Google", Host: "google.com"},
	{Name: "Cloudflare", Host: "1.1.1.1"},
	{Name: "Apple", Host: "apple.com"},
	{Name: "Amazon", Host: "amazon.com"},
	{Name: "GitHub", Host: "github.com"},
	{Name: "Quad9", Host: "9.9.9.9"},
	{Name: "OpenDNS", Host: "208.67.222.222"},
}

// SpeedEndpoint is a preconfigured speed test server.
type SpeedEndpoint struct {
	Name      string `json:"name"`
	URL       string `json:"url,omitempty"`
	UploadURL string `json:"upload_url,omitempty"` // empty if upload not supported
	Bytes     int    `json:"bytes,omitempty"`       // download size in bytes
	Ookla     bool   `json:"ookla,omitempty"`       // use Ookla speedtest CLI instead of curl
}

// SpeedEndpoints provides preconfigured endpoints for speed tests.
// The Ookla entry is only usable when `speedtest` CLI is installed.
var SpeedEndpoints = []SpeedEndpoint{
	{Name: "Quick (10MB)", URL: "https://speed.cloudflare.com/__down?bytes=10000000", UploadURL: "https://speed.cloudflare.com/__up", Bytes: 10_000_000},
	{Name: "Standard (25MB)", URL: "https://speed.cloudflare.com/__down?bytes=25000000", UploadURL: "https://speed.cloudflare.com/__up", Bytes: 25_000_000},
	{Name: "Thorough (100MB)", URL: "https://speed.cloudflare.com/__down?bytes=100000000", UploadURL: "https://speed.cloudflare.com/__up", Bytes: 100_000_000},
	{Name: "Ookla (auto server)", Ookla: true},
}

// SpeedHistoryEntry is a timestamped speed test result for history tracking.
type SpeedHistoryEntry struct {
	DownloadMbps float64   `json:"download_mbps"`
	UploadMbps   float64   `json:"upload_mbps"`
	LatencyMs    float64   `json:"latency_ms"`
	Server       string    `json:"server"`
	Timestamp    time.Time `json:"timestamp"`
}

// DNSBenchmarkResult holds the result of a DNS benchmark test.
type DNSBenchmarkResult struct {
	Provider   string  `json:"provider"`
	Server     string  `json:"server"`
	LatencyMs  float64 `json:"latency_ms"`
	Error      string  `json:"error,omitempty"`
	Reachable  bool    `json:"reachable"`
}

// DNSBenchmarkReport holds the full benchmark report.
type DNSBenchmarkReport struct {
	Results     []DNSBenchmarkResult `json:"results"`
	Fastest     string               `json:"fastest"`
	FastestMs   float64              `json:"fastest_ms"`
	TestedAt    time.Time            `json:"tested_at"`
}

// DiagnoseCheck represents a single diagnostic check result.
type DiagnoseCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "ok", "warning", "error"
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

// DiagnoseReport holds the full diagnostic report.
type DiagnoseReport struct {
	Checks    []DiagnoseCheck `json:"checks"`
	Summary   string          `json:"summary"`
	Timestamp time.Time       `json:"timestamp"`
}

// WiFiSignalEvent represents a single WiFi signal measurement.
type WiFiSignalEvent struct {
	SSID      string    `json:"ssid"`
	RSSI      int       `json:"rssi"`
	Noise     int       `json:"noise"`
	SNR       int       `json:"snr"`
	Quality   string    `json:"quality"`
	Timestamp time.Time `json:"timestamp"`
}
