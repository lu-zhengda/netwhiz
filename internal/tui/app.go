package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lu-zhengda/netwhiz/internal/network"
)

// Tab indices.
const (
	tabOverview = iota
	tabWiFi
	tabDNS
	tabPing
	tabSpeed
	tabCount
)

var tabNames = []string{"Overview", "WiFi", "DNS", "Ping", "Speed"}

// Messages.
type networkInfoMsg struct {
	info *network.NetworkInfo
	err  error
}

type wifiInfoMsg struct {
	info *network.WiFiInfo
	err  error
}

type wifiScanMsg struct {
	networks []network.WiFiNetwork
	err      error
}

type dnsInfoMsg struct {
	servers     []string
	serviceName string
	err         error
}

type pingResultMsg struct {
	stats *network.PingStats
	err   error
}

type speedResultMsg struct {
	result *network.SpeedResult
	err    error
}

type dnsSetMsg struct {
	preset string
	err    error
}

type tickMsg time.Time

// Model is the main TUI model.
type Model struct {
	runner  network.CmdRunner
	version string

	activeTab int
	width     int
	height    int

	spinner spinner.Model

	// Overview tab.
	networkInfo *network.NetworkInfo
	infoLoading bool
	infoErr     error

	// WiFi tab.
	wifiInfo      *network.WiFiInfo
	wifiNetworks  []network.WiFiNetwork
	wifiLoading   bool
	wifiScanning  bool
	wifiErr       error
	wifiScanErr   error

	// DNS tab.
	dnsServers     []string
	dnsServiceName string
	dnsPresetIdx   int
	dnsLoading     bool
	dnsSetting     bool // true while applying a preset
	dnsErr         error
	dnsSetErr      error // error from last preset apply

	// Ping tab.
	pingTargetIdx int // index into network.PingTargets
	pingHost      string
	pingInput     textinput.Model
	pingStats     *network.PingStats
	pingHistory   []time.Duration // recent ping times for sparkline
	pingRunning   bool
	pingErr       error

	// Speed tab.
	speedEndpointIdx int // index into network.SpeedEndpoints
	speedResult      *network.SpeedResult
	speedRunning     bool
	speedErr         error
}

// New creates a new TUI model.
func New(runner network.CmdRunner, version string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	ti := textinput.New()
	ti.Placeholder = "google.com"
	ti.CharLimit = 256
	ti.Width = 40

	return Model{
		runner:  runner,
		version: version,
		spinner: sp,

		pingHost:  "google.com",
		pingInput: ti,
	}
}

// Init starts the TUI by loading initial data.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadOverviewData(),
		m.loadWiFiData(),
		m.loadDNSData(),
	)
}

// loadOverviewData fetches network info in the background.
func (m Model) loadOverviewData() tea.Cmd {
	runner := m.runner
	return func() tea.Msg {
		svc := network.NewInfoService(runner)
		info, err := svc.GetNetworkInfo(context.Background())
		return networkInfoMsg{info: info, err: err}
	}
}

// loadWiFiData fetches WiFi info in the background.
func (m Model) loadWiFiData() tea.Cmd {
	runner := m.runner
	return func() tea.Msg {
		svc := network.NewWiFiService(runner)
		info, err := svc.GetWiFiInfo(context.Background())
		return wifiInfoMsg{info: info, err: err}
	}
}

// scanWiFiNetworks scans for nearby WiFi networks.
func (m Model) scanWiFiNetworks() tea.Cmd {
	runner := m.runner
	return func() tea.Msg {
		svc := network.NewWiFiService(runner)
		networks, err := svc.ScanNetworks(context.Background())
		return wifiScanMsg{networks: networks, err: err}
	}
}

// loadDNSData fetches DNS info in the background.
func (m Model) loadDNSData() tea.Cmd {
	runner := m.runner
	return func() tea.Msg {
		svc := network.NewDNSService(runner)
		servers, serviceName, err := svc.GetDNSServers(context.Background())
		return dnsInfoMsg{servers: servers, serviceName: serviceName, err: err}
	}
}

// runPing executes a ping in the background.
func (m Model) runPing() tea.Cmd {
	runner := m.runner
	host := m.pingHost
	return func() tea.Msg {
		svc := network.NewPingService(runner)
		stats, err := svc.Ping(context.Background(), host, 10)
		return pingResultMsg{stats: stats, err: err}
	}
}

// setDNSPreset applies a DNS preset in the background.
func (m Model) setDNSPreset() tea.Cmd {
	runner := m.runner
	preset := network.DNSPresets[m.dnsPresetIdx]
	return func() tea.Msg {
		svc := network.NewDNSService(runner)
		err := svc.SetDNS(context.Background(), preset.Name)
		return dnsSetMsg{preset: preset.Name, err: err}
	}
}

// runSpeedTest executes a speed test in the background.
func (m Model) runSpeedTest() tea.Cmd {
	runner := m.runner
	endpoint := network.SpeedEndpoints[m.speedEndpointIdx]
	return func() tea.Msg {
		svc := network.NewSpeedService(runner)
		result, err := svc.RunSpeedTestWith(context.Background(), endpoint)
		return speedResultMsg{result: result, err: err}
	}
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		if m.infoLoading || m.wifiLoading || m.wifiScanning || m.dnsLoading || m.dnsSetting || m.pingRunning || m.speedRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case networkInfoMsg:
		m.infoLoading = false
		m.networkInfo = msg.info
		m.infoErr = msg.err
		return m, nil

	case wifiInfoMsg:
		m.wifiLoading = false
		m.wifiInfo = msg.info
		m.wifiErr = msg.err
		return m, nil

	case wifiScanMsg:
		m.wifiScanning = false
		m.wifiNetworks = msg.networks
		m.wifiScanErr = msg.err
		return m, nil

	case dnsInfoMsg:
		m.dnsLoading = false
		m.dnsServers = msg.servers
		m.dnsServiceName = msg.serviceName
		m.dnsErr = msg.err
		return m, nil

	case dnsSetMsg:
		m.dnsSetting = false
		m.dnsSetErr = msg.err
		// Reload DNS info to show updated servers.
		m.dnsLoading = true
		return m, tea.Batch(m.loadDNSData(), m.spinner.Tick)

	case pingResultMsg:
		m.pingRunning = false
		m.pingStats = msg.stats
		m.pingErr = msg.err
		if msg.stats != nil {
			for _, r := range msg.stats.Results {
				m.pingHistory = append(m.pingHistory, r.Time)
			}
			// Keep last 50 entries.
			if len(m.pingHistory) > 50 {
				m.pingHistory = m.pingHistory[len(m.pingHistory)-50:]
			}
		}
		return m, nil

	case speedResultMsg:
		m.speedRunning = false
		m.speedResult = msg.result
		m.speedErr = msg.err
		return m, nil
	}

	// Update text input if on ping tab.
	if m.activeTab == tabPing && m.pingInput.Focused() {
		var cmd tea.Cmd
		m.pingInput, cmd = m.pingInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If ping input is focused, handle that first.
	if m.activeTab == tabPing && m.pingInput.Focused() {
		switch msg.String() {
		case "enter":
			host := m.pingInput.Value()
			if host == "" {
				host = "google.com"
			}
			m.pingHost = host
			m.pingInput.Blur()
			m.pingRunning = true
			m.pingErr = nil
			return m, tea.Batch(m.runPing(), m.spinner.Tick)
		case "esc":
			m.pingInput.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.pingInput, cmd = m.pingInput.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "tab", "right", "l":
		m.activeTab = (m.activeTab + 1) % tabCount
		return m, nil

	case "shift+tab", "left", "h":
		m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
		return m, nil

	case "1":
		m.activeTab = tabOverview
		return m, nil
	case "2":
		m.activeTab = tabWiFi
		return m, nil
	case "3":
		m.activeTab = tabDNS
		return m, nil
	case "4":
		m.activeTab = tabPing
		return m, nil
	case "5":
		m.activeTab = tabSpeed
		return m, nil

	case "r":
		return m.handleRefresh()

	case "p":
		if m.activeTab != tabPing {
			m.activeTab = tabPing
		}
		if !m.pingRunning {
			m.pingRunning = true
			m.pingErr = nil
			return m, tea.Batch(m.runPing(), m.spinner.Tick)
		}
		return m, nil

	case "d":
		if m.activeTab != tabDNS {
			m.activeTab = tabDNS
		}
		return m, nil

	case "s":
		if m.activeTab != tabSpeed {
			m.activeTab = tabSpeed
		}
		if !m.speedRunning {
			m.speedRunning = true
			m.speedErr = nil
			return m, tea.Batch(m.runSpeedTest(), m.spinner.Tick)
		}
		return m, nil

	case "w":
		if m.activeTab == tabWiFi && !m.wifiScanning {
			m.wifiScanning = true
			return m, tea.Batch(m.scanWiFiNetworks(), m.spinner.Tick)
		}
		return m, nil

	case "i":
		if m.activeTab == tabPing && !m.pingInput.Focused() {
			m.pingInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "down", "j", "n":
		// Move to next target/endpoint/preset.
		if m.activeTab == tabDNS && !m.dnsSetting {
			m.dnsPresetIdx = (m.dnsPresetIdx + 1) % len(network.DNSPresets)
			return m, nil
		}
		if m.activeTab == tabPing && !m.pingRunning {
			m.pingTargetIdx = (m.pingTargetIdx + 1) % len(network.PingTargets)
			m.pingHost = network.PingTargets[m.pingTargetIdx].Host
			m.pingStats = nil
			m.pingErr = nil
			return m, nil
		}
		if m.activeTab == tabSpeed && !m.speedRunning {
			m.speedEndpointIdx = (m.speedEndpointIdx + 1) % len(network.SpeedEndpoints)
			m.speedResult = nil
			m.speedErr = nil
			return m, nil
		}
		return m, nil

	case "up", "k", "N":
		// Move to previous target/endpoint/preset.
		if m.activeTab == tabDNS && !m.dnsSetting {
			m.dnsPresetIdx = (m.dnsPresetIdx - 1 + len(network.DNSPresets)) % len(network.DNSPresets)
			return m, nil
		}
		if m.activeTab == tabPing && !m.pingRunning {
			m.pingTargetIdx = (m.pingTargetIdx - 1 + len(network.PingTargets)) % len(network.PingTargets)
			m.pingHost = network.PingTargets[m.pingTargetIdx].Host
			m.pingStats = nil
			m.pingErr = nil
			return m, nil
		}
		if m.activeTab == tabSpeed && !m.speedRunning {
			m.speedEndpointIdx = (m.speedEndpointIdx - 1 + len(network.SpeedEndpoints)) % len(network.SpeedEndpoints)
			m.speedResult = nil
			m.speedErr = nil
			return m, nil
		}
		return m, nil

	case "enter":
		if m.activeTab == tabDNS && !m.dnsSetting {
			m.dnsSetting = true
			m.dnsSetErr = nil
			return m, tea.Batch(m.setDNSPreset(), m.spinner.Tick)
		}
		if m.activeTab == tabPing && !m.pingRunning {
			m.pingRunning = true
			m.pingErr = nil
			return m, tea.Batch(m.runPing(), m.spinner.Tick)
		}
		if m.activeTab == tabSpeed && !m.speedRunning {
			m.speedRunning = true
			m.speedErr = nil
			return m, tea.Batch(m.runSpeedTest(), m.spinner.Tick)
		}
		return m, nil
	}

	return m, nil
}

// handleRefresh reloads data for the current tab.
func (m Model) handleRefresh() (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch m.activeTab {
	case tabOverview:
		m.infoLoading = true
		m.infoErr = nil
		cmds = append(cmds, m.loadOverviewData())
	case tabWiFi:
		m.wifiLoading = true
		m.wifiErr = nil
		cmds = append(cmds, m.loadWiFiData())
	case tabDNS:
		m.dnsLoading = true
		m.dnsErr = nil
		cmds = append(cmds, m.loadDNSData())
	case tabPing:
		if !m.pingRunning {
			m.pingRunning = true
			m.pingErr = nil
			cmds = append(cmds, m.runPing())
		}
	case tabSpeed:
		if !m.speedRunning {
			m.speedRunning = true
			m.speedErr = nil
			cmds = append(cmds, m.runSpeedTest())
		}
	}

	cmds = append(cmds, m.spinner.Tick)
	return m, tea.Batch(cmds...)
}

// View renders the TUI.
func (m Model) View() string {
	var b strings.Builder

	// Header.
	header := headerStyle.Render(" netwhiz ")
	b.WriteString(header)
	b.WriteString("\n")

	// Tab bar.
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Separator.
	if m.width > 0 {
		b.WriteString(strings.Repeat("\u2500", m.width))
	} else {
		b.WriteString(strings.Repeat("\u2500", 70))
	}
	b.WriteString("\n")

	// Tab content.
	switch m.activeTab {
	case tabOverview:
		b.WriteString(m.viewOverview())
	case tabWiFi:
		b.WriteString(m.viewWiFi())
	case tabDNS:
		b.WriteString(m.viewDNS())
	case tabPing:
		b.WriteString(m.viewPing())
	case tabSpeed:
		b.WriteString(m.viewSpeed())
	}

	// Help bar.
	b.WriteString("\n")
	if m.width > 0 {
		b.WriteString(strings.Repeat("\u2500", m.width))
	} else {
		b.WriteString(strings.Repeat("\u2500", 70))
	}
	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return b.String()
}

// renderTabs renders the tab bar.
func (m Model) renderTabs() string {
	var tabs []string
	for i, name := range tabNames {
		if i == m.activeTab {
			tabs = append(tabs, activeTabStyle.Render("["+name+"]"))
		} else {
			tabs = append(tabs, tabStyle.Render(" "+name+" "))
		}
	}
	return strings.Join(tabs, " ")
}

// renderHelp renders the bottom help bar.
func (m Model) renderHelp() string {
	parts := []string{
		"Tab/\u2190\u2192:switch",
		"p:ping",
		"d:dns",
		"s:speed",
		"r:refresh",
		"q:quit",
	}

	if m.activeTab == tabWiFi {
		parts = append([]string{"w:scan networks"}, parts...)
	}
	if m.activeTab == tabDNS {
		parts = append([]string{"j/k:select preset", "enter:apply"}, parts...)
	}
	if m.activeTab == tabPing {
		parts = append([]string{"i:set host", "j/k:select target", "enter:run"}, parts...)
	}
	if m.activeTab == tabSpeed {
		parts = append([]string{"j/k:select server", "enter:run"}, parts...)
	}

	return helpStyle.Render(strings.Join(parts, "  "))
}

// --- Tab Views ---

func (m Model) viewOverview() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("  Network Overview"))
	b.WriteString("\n")
	b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

	if m.infoLoading {
		b.WriteString("  " + m.spinner.View() + " Loading network info...\n")
		return b.String()
	}

	if m.infoErr != nil {
		b.WriteString("  " + errorStyle.Render("Error: "+m.infoErr.Error()) + "\n")
		return b.String()
	}

	if m.networkInfo != nil {
		info := m.networkInfo
		b.WriteString(fmt.Sprintf("  %s  %s              %s  %s\n",
			labelStyle.Render("Interface:"),
			valueStyle.Render(info.Interface),
			labelStyle.Render("Status:"),
			colorStatus(info.Status)))
		b.WriteString(fmt.Sprintf("  %s %s       %s  %s\n",
			labelStyle.Render("IP Address:"),
			valueStyle.Render(info.IPAddress),
			labelStyle.Render("Router:"),
			valueStyle.Render(info.Router)))

		dns := "(auto)"
		if len(info.DNS) > 0 {
			dns = strings.Join(info.DNS, ", ")
		}
		publicIP := info.PublicIP
		if publicIP == "" {
			publicIP = "(unavailable)"
		}
		b.WriteString(fmt.Sprintf("  %s        %s  %s  %s\n",
			labelStyle.Render("DNS:"),
			valueStyle.Render(dns),
			labelStyle.Render("Public:"),
			valueStyle.Render(publicIP)))
	}

	// WiFi summary.
	if m.wifiInfo != nil {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render("  WiFi"))
		b.WriteString("\n")
		b.WriteString("  \u2550\u2550\u2550\u2550\n")

		wifi := m.wifiInfo
		signalBar := network.SignalBar(wifi.RSSI)
		b.WriteString(fmt.Sprintf("  SSID: %s  Channel: %d (%s)  Signal: %ddBm (%s)\n",
			valueStyle.Render(wifi.SSID),
			wifi.Channel,
			wifi.Band,
			wifi.RSSI,
			signalBar))
		b.WriteString(fmt.Sprintf("  Tx Rate: %s  Security: %s\n",
			valueStyle.Render(fmt.Sprintf("%d Mbps", wifi.TxRate)),
			valueStyle.Render(wifi.Security)))
	}

	// Ping summary.
	if m.pingStats != nil {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render(fmt.Sprintf("  Recent Ping (%s)", m.pingHost)))
		b.WriteString("\n")
		b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

		sparkline := renderSparkline(m.pingHistory)
		b.WriteString(fmt.Sprintf("  %s  Avg: %.1fms  Loss: %.0f%%\n",
			sparkline,
			float64(m.pingStats.AvgRTT)/float64(time.Millisecond),
			m.pingStats.LossPercent))
	}

	// Speed summary.
	if m.speedResult != nil {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render("  Speed Test"))
		b.WriteString("\n")
		b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")
		upload := ""
		if m.speedResult.UploadMbps > 0 {
			upload = fmt.Sprintf("  Upload: %s", valueStyle.Render(network.FormatSpeed(m.speedResult.UploadMbps)))
		}
		b.WriteString(fmt.Sprintf("  Download: %s%s  Latency: %.0fms\n",
			valueStyle.Render(network.FormatSpeed(m.speedResult.DownloadMbps)),
			upload,
			float64(m.speedResult.Latency)/float64(time.Millisecond)))
	}

	return b.String()
}

func (m Model) viewWiFi() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("  WiFi Information"))
	b.WriteString("\n")
	b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

	if m.wifiLoading {
		b.WriteString("  " + m.spinner.View() + " Loading WiFi info...\n")
		return b.String()
	}

	if m.wifiErr != nil {
		b.WriteString("  " + errorStyle.Render("Error: "+m.wifiErr.Error()) + "\n")
		return b.String()
	}

	if m.wifiInfo != nil {
		wifi := m.wifiInfo
		b.WriteString(fmt.Sprintf("  %s        %s\n", labelStyle.Render("SSID:"), valueStyle.Render(wifi.SSID)))
		b.WriteString(fmt.Sprintf("  %s       %s\n", labelStyle.Render("BSSID:"), valueStyle.Render(wifi.BSSID)))
		b.WriteString(fmt.Sprintf("  %s     %s (%s)\n", labelStyle.Render("Channel:"), valueStyle.Render(fmt.Sprintf("%d", wifi.Channel)), wifi.Band))
		b.WriteString(fmt.Sprintf("  %s        %s dBm (%s)\n", labelStyle.Render("RSSI:"),
			valueStyle.Render(fmt.Sprintf("%d", wifi.RSSI)),
			network.SignalQuality(wifi.RSSI)))
		b.WriteString(fmt.Sprintf("  %s       %s dBm\n", labelStyle.Render("Noise:"), valueStyle.Render(fmt.Sprintf("%d", wifi.Noise))))
		b.WriteString(fmt.Sprintf("  %s         %s dB\n", labelStyle.Render("SNR:"), valueStyle.Render(fmt.Sprintf("%d", wifi.SNR))))
		b.WriteString(fmt.Sprintf("  %s     %s Mbps\n", labelStyle.Render("Tx Rate:"), valueStyle.Render(fmt.Sprintf("%d", wifi.TxRate))))
		b.WriteString(fmt.Sprintf("  %s    %s\n", labelStyle.Render("Security:"), valueStyle.Render(wifi.Security)))
		b.WriteString(fmt.Sprintf("  %s     %s\n", labelStyle.Render("Country:"), valueStyle.Render(wifi.CountryCode)))
	}

	// Nearby networks.
	if m.wifiScanning {
		b.WriteString("\n  " + m.spinner.View() + " Scanning nearby networks...\n")
	} else if m.wifiScanErr != nil {
		b.WriteString("\n  " + errorStyle.Render("Scan error: "+m.wifiScanErr.Error()) + "\n")
	} else if len(m.wifiNetworks) > 0 {
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render("  Nearby Networks"))
		b.WriteString("\n")
		b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

		b.WriteString(fmt.Sprintf("  %-25s  %6s  %4s  %s\n", "SSID", "RSSI", "CH", "SECURITY"))

		maxShow := len(m.wifiNetworks)
		if maxShow > 15 {
			maxShow = 15
		}
		for _, n := range m.wifiNetworks[:maxShow] {
			signal := network.SignalBar(n.RSSI)
			ssid := n.SSID
			if len(ssid) > 25 {
				ssid = ssid[:22] + "..."
			}
			b.WriteString(fmt.Sprintf("  %-25s  %4ddBm  %4d  %s  %s\n",
				ssid, n.RSSI, n.Channel, signal, n.Security))
		}

		if len(m.wifiNetworks) > 15 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("\n  ... and %d more networks", len(m.wifiNetworks)-15)) + "\n")
		}
	}

	return b.String()
}

func (m Model) viewDNS() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("  DNS Configuration"))
	b.WriteString("\n")
	b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

	if m.dnsLoading {
		b.WriteString("  " + m.spinner.View() + " Loading DNS info...\n")
		return b.String()
	}

	if m.dnsErr != nil {
		b.WriteString("  " + errorStyle.Render("Error: "+m.dnsErr.Error()) + "\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("  %s  %s\n",
		labelStyle.Render("Service:"),
		valueStyle.Render(m.dnsServiceName)))

	if len(m.dnsServers) > 0 {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			labelStyle.Render("Servers:"),
			valueStyle.Render(strings.Join(m.dnsServers, ", "))))

		// Identify preset.
		for _, p := range network.DNSPresets {
			if matchesPreset(m.dnsServers, p.Servers) {
				b.WriteString(fmt.Sprintf("  %s   %s\n",
					labelStyle.Render("Preset:"),
					valueStyle.Render(p.Name)))
				break
			}
		}
	} else {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			labelStyle.Render("Servers:"),
			dimStyle.Render("(auto / DHCP)")))
	}

	if m.dnsSetErr != nil {
		b.WriteString("\n  " + errorStyle.Render("Set failed: "+m.dnsSetErr.Error()))
		b.WriteString("\n  " + dimStyle.Render("(may need sudo: sudo netwhiz dns set <preset>)") + "\n")
	}

	// Available presets.
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("  Available Presets"))
	b.WriteString("\n")
	b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

	if m.dnsSetting {
		b.WriteString("  " + m.spinner.View() + " Applying preset...\n")
	}

	for i, p := range network.DNSPresets {
		marker := "  "
		if i == m.dnsPresetIdx {
			marker = "> "
		}
		active := ""
		if matchesPreset(m.dnsServers, p.Servers) {
			active = " " + successStyle.Render("(active)")
		}
		name := fmt.Sprintf("%-12s  %s%s", p.Name, strings.Join(p.Servers, ", "), active)
		if i == m.dnsPresetIdx {
			b.WriteString(selectedStyle.Render(marker+name) + "\n")
		} else {
			b.WriteString(marker + name + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  j/k:select  enter:apply  (may require sudo)"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewPing() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render(fmt.Sprintf("  Ping %s", m.pingHost)))
	b.WriteString("\n")
	b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

	if m.pingInput.Focused() {
		b.WriteString("\n  Host: " + m.pingInput.View() + "\n")
		b.WriteString(dimStyle.Render("  Press Enter to ping, Esc to cancel") + "\n")
		return b.String()
	}

	if m.pingRunning {
		b.WriteString("  " + m.spinner.View() + " Pinging " + m.pingHost + "...\n")
	}

	if m.pingErr != nil {
		b.WriteString("  " + errorStyle.Render("Error: "+m.pingErr.Error()) + "\n")
	}

	if m.pingStats != nil {
		stats := m.pingStats

		// Find max RTT for bar scaling.
		var maxRTT time.Duration
		for _, r := range stats.Results {
			if r.Time > maxRTT {
				maxRTT = r.Time
			}
		}

		// Individual results with bars.
		for _, r := range stats.Results {
			bar := network.PingBar(r.Time, maxRTT, 20)
			b.WriteString(fmt.Sprintf("  %2d: %6.1fms  %s\n",
				r.Seq+1,
				float64(r.Time)/float64(time.Millisecond),
				bar))
		}

		// Statistics.
		b.WriteString("\n")
		b.WriteString(sectionStyle.Render("  Statistics"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Sent: %d  Received: %d  Lost: %d (%.0f%%)\n",
			stats.Sent, stats.Received, stats.Lost, stats.LossPercent))
		b.WriteString(fmt.Sprintf("  Min: %.1fms  Avg: %.1fms  Max: %.1fms  StdDev: %.1fms\n",
			float64(stats.MinRTT)/float64(time.Millisecond),
			float64(stats.AvgRTT)/float64(time.Millisecond),
			float64(stats.MaxRTT)/float64(time.Millisecond),
			float64(stats.StdDevRTT)/float64(time.Millisecond)))

		// Sparkline of history.
		if len(m.pingHistory) > 0 {
			b.WriteString("\n")
			b.WriteString(sectionStyle.Render("  History"))
			b.WriteString("\n")
			b.WriteString("  " + renderSparkline(m.pingHistory) + "\n")
		}
	} else if !m.pingRunning {
		b.WriteString("\n  Press 'p' or Enter to start pinging\n")
		b.WriteString("  Press 'i' to change host, j/k to select target\n")
	}

	// Show available targets.
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("  Targets"))
	b.WriteString("\n")
	for i, t := range network.PingTargets {
		marker := "  "
		if t.Host == m.pingHost {
			marker = "> "
		}
		name := fmt.Sprintf("%-12s %s", t.Name, dimStyle.Render(t.Host))
		if i == m.pingTargetIdx {
			b.WriteString(selectedStyle.Render(marker + name) + "\n")
		} else {
			b.WriteString(marker + name + "\n")
		}
	}

	return b.String()
}

func (m Model) viewSpeed() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("  Speed Test"))
	b.WriteString("\n")
	b.WriteString("  \u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")

	if m.speedRunning {
		ep := network.SpeedEndpoints[m.speedEndpointIdx]
		msg := " Testing speed (" + ep.Name + ")..."
		if ep.UploadURL != "" {
			msg = " Testing download + upload (" + ep.Name + ")..."
		}
		b.WriteString("  " + m.spinner.View() + msg + "\n")
		b.WriteString(dimStyle.Render("  This may take up to 30 seconds...") + "\n")
		return b.String()
	}

	if m.speedErr != nil {
		b.WriteString("  " + errorStyle.Render("Error: "+m.speedErr.Error()) + "\n")
		b.WriteString("\n  Press 's' or Enter to retry, j/k to change server\n")
	}

	if m.speedResult != nil {
		result := m.speedResult

		// Download speed with a visual bar.
		dlBar := renderSpeedBar(result.DownloadMbps)
		b.WriteString(fmt.Sprintf("\n  %s  %s\n",
			labelStyle.Render("Download:"),
			valueStyle.Render(network.FormatSpeed(result.DownloadMbps))))
		b.WriteString(fmt.Sprintf("  %s\n", dlBar))

		// Upload speed (if measured).
		if result.UploadMbps > 0 {
			ulBar := renderSpeedBar(result.UploadMbps)
			b.WriteString(fmt.Sprintf("\n  %s    %s\n",
				labelStyle.Render("Upload:"),
				valueStyle.Render(network.FormatSpeed(result.UploadMbps))))
			b.WriteString(fmt.Sprintf("  %s\n", ulBar))
		}

		b.WriteString(fmt.Sprintf("\n  %s   %.0fms\n",
			labelStyle.Render("Latency:"),
			float64(result.Latency)/float64(time.Millisecond)))
		b.WriteString(fmt.Sprintf("  %s    %s\n",
			labelStyle.Render("Server:"),
			valueStyle.Render(result.Server)))
		b.WriteString(fmt.Sprintf("  %s      %s\n",
			labelStyle.Render("Time:"),
			dimStyle.Render(result.Timestamp.Format("2006-01-02 15:04:05"))))

		b.WriteString("\n  Press 's' or Enter to test again, j/k to change server\n")
	} else if m.speedErr == nil {
		b.WriteString("\n  Press 's' or Enter to start speed test, j/k to change server\n")
	}

	// Show available endpoints.
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("  Servers"))
	b.WriteString("\n")
	ooklaInstalled := network.OoklaAvailable()
	for i, ep := range network.SpeedEndpoints {
		marker := "  "
		if i == m.speedEndpointIdx {
			marker = "> "
		}
		name := ep.Name
		if ep.Ookla && !ooklaInstalled {
			name += " " + dimStyle.Render("(brew tap teamookla/speedtest && brew install speedtest)")
		}
		if i == m.speedEndpointIdx {
			b.WriteString(selectedStyle.Render(marker+name) + "\n")
		} else {
			b.WriteString(marker + name + "\n")
		}
	}

	return b.String()
}

// --- Helper functions ---

// renderSparkline renders a sparkline chart from a series of durations.
func renderSparkline(values []time.Duration) string {
	if len(values) == 0 {
		return ""
	}

	// Find min and max.
	minVal := values[0]
	maxVal := values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	rangeVal := maxVal - minVal
	if rangeVal == 0 {
		rangeVal = 1
	}

	var sb strings.Builder
	for _, v := range values {
		idx := int(float64(v-minVal) / float64(rangeVal) * float64(len(sparklineChars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparklineChars) {
			idx = len(sparklineChars) - 1
		}
		sb.WriteRune(sparklineChars[idx])
	}

	return sb.String()
}

// renderSpeedBar renders a visual bar for download speed.
func renderSpeedBar(mbps float64) string {
	// Scale: 0-1000 Mbps.
	maxMbps := 1000.0
	if mbps > maxMbps {
		maxMbps = mbps
	}

	width := 40
	filled := int(mbps / maxMbps * float64(width))
	if filled == 0 && mbps > 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", width-filled)
	return "  [" + bar + "]"
}

// colorStatus returns a styled status string.
func colorStatus(status string) string {
	switch status {
	case "Active":
		return successStyle.Render(status)
	case "Inactive":
		return errorStyle.Render(status)
	default:
		return warnStyle.Render(status)
	}
}

// matchesPreset checks if the given servers match a DNS preset.
func matchesPreset(servers, preset []string) bool {
	if len(servers) != len(preset) {
		return false
	}
	for i := range servers {
		if servers[i] != preset[i] {
			return false
		}
	}
	return true
}
