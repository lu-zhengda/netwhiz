package network

import "time"

// NetworkInfo holds system network interface information.
type NetworkInfo struct {
	Interface  string
	IPAddress  string
	SubnetMask string
	Router     string
	DNS        []string
	MACAddress string
	MediaSpeed string
	Status     string // "active", "inactive"
	PublicIP   string
}

// WiFiInfo holds current WiFi connection details.
type WiFiInfo struct {
	SSID        string
	BSSID       string
	Channel     int
	Band        string // "2.4GHz", "5GHz", "6GHz"
	RSSI        int    // signal strength in dBm
	Noise       int    // noise in dBm
	SNR         int    // signal-to-noise ratio
	TxRate      int    // Mbps
	Security    string // "WPA2", "WPA3", etc.
	CountryCode string
}

// WiFiNetwork represents a nearby WiFi network from scanning.
type WiFiNetwork struct {
	SSID     string
	BSSID    string
	RSSI     int
	Channel  int
	Security string
}

// PingResult holds a single ping response.
type PingResult struct {
	Host      string
	Seq       int
	TTL       int
	Time      time.Duration
	Timestamp time.Time
}

// PingStats holds aggregate ping statistics.
type PingStats struct {
	Host        string
	Sent        int
	Received    int
	Lost        int
	LossPercent float64
	MinRTT      time.Duration
	MaxRTT      time.Duration
	AvgRTT      time.Duration
	StdDevRTT   time.Duration
	Results     []PingResult
}

// TraceHop holds a single traceroute hop.
type TraceHop struct {
	Hop      int
	IP       string
	Hostname string
	RTTs     []time.Duration // up to 3 probes
	Lost     bool
}

// SpeedResult holds speed test results.
type SpeedResult struct {
	DownloadMbps float64
	UploadMbps   float64
	Latency      time.Duration
	Server       string
	Timestamp    time.Time
}

// ARPEntry holds a single ARP table entry.
type ARPEntry struct {
	IP        string
	MAC       string
	Interface string
	Hostname  string // reverse DNS if available
}

// DNSPreset holds a friendly name and DNS server addresses.
type DNSPreset struct {
	Name    string
	Servers []string
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
	Name string
	Host string
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
	Name      string
	URL       string
	UploadURL string // empty if upload not supported
	Bytes     int    // download size in bytes
	Ookla     bool   // use Ookla speedtest CLI instead of curl
}

// SpeedEndpoints provides preconfigured endpoints for speed tests.
// The Ookla entry is only usable when `speedtest` CLI is installed.
var SpeedEndpoints = []SpeedEndpoint{
	{Name: "Quick (10MB)", URL: "https://speed.cloudflare.com/__down?bytes=10000000", UploadURL: "https://speed.cloudflare.com/__up", Bytes: 10_000_000},
	{Name: "Standard (25MB)", URL: "https://speed.cloudflare.com/__down?bytes=25000000", UploadURL: "https://speed.cloudflare.com/__up", Bytes: 25_000_000},
	{Name: "Thorough (100MB)", URL: "https://speed.cloudflare.com/__down?bytes=100000000", UploadURL: "https://speed.cloudflare.com/__up", Bytes: 100_000_000},
	{Name: "Ookla (auto server)", Ookla: true},
}
