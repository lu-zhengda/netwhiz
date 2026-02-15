# netwhiz

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform: macOS](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://github.com/lu-zhengda/netwhiz)
[![Homebrew](https://img.shields.io/badge/Homebrew-lu--zhengda/tap-orange.svg)](https://github.com/lu-zhengda/homebrew-tap)

Network diagnostics toolkit for macOS — inspect interfaces, test connectivity, manage DNS, run speed tests, and scan your LAN.

## Install

```bash
brew tap lu-zhengda/tap
brew install netwhiz
```

## Usage

```
$ netwhiz info
Interface: en0 (Wi-Fi)    Status: active
Local IP:  192.168.1.42   Public IP: 203.0.113.1
Gateway:   192.168.1.1    DNS: 1.1.1.1, 1.0.0.1

$ netwhiz wifi
SSID:     MyNetwork
Channel:  36 (5 GHz)
RSSI:     -42 dBm (Excellent)
SNR:      35 dB
Security: WPA3 Personal

$ netwhiz speed
Download: 245.3 Mbps (via Cloudflare)
```

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `info` | Network overview (IP, gateway, DNS, interface, public IP) | `netwhiz info` |
| `wifi` | WiFi details (SSID, channel, RSSI, SNR, security) | `netwhiz wifi` |
| `wifi scan` | Scan nearby WiFi networks with signal bars | `netwhiz wifi scan` |
| `dns` | Show current DNS servers and presets | `netwhiz dns` |
| `dns set <server>` | Set DNS (cloudflare, google, quad9, or custom IP) | `netwhiz dns set cloudflare` |
| `dns flush` | Flush DNS cache | `netwhiz dns flush` |
| `ping <host>` | Enhanced ping with stats and RTT bars | `netwhiz ping 8.8.8.8 -c 10` |
| `trace <host>` | Traceroute with per-hop RTT | `netwhiz trace example.com` |
| `speed` | Download speed test via Cloudflare | `netwhiz speed` |
| `scan` | ARP scan to discover LAN devices | `netwhiz scan` |

### DNS Presets

| Preset | Servers |
|--------|---------|
| `cloudflare` | 1.1.1.1, 1.0.0.1 |
| `google` | 8.8.8.8, 8.8.4.4 |
| `quad9` | 9.9.9.9, 149.112.112.112 |

## Diagnostic Workflow

When troubleshooting network issues, follow this order:

1. `netwhiz info` — get network overview and identify interface
2. `netwhiz wifi` — check signal quality (if wireless)
3. `netwhiz dns` — verify DNS configuration
4. `netwhiz ping <target>` — test connectivity and latency
5. `netwhiz trace <target>` — identify where packets are dropping
6. `netwhiz speed` — measure throughput

## TUI

Launch `netwhiz` without arguments for an interactive network dashboard. Run diagnostics, view results, and navigate between tools with a keyboard-driven interface.

## License

[MIT](LICENSE)
