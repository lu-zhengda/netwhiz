# netwhiz

A network diagnostics toolkit for macOS — inspect interfaces, test connectivity, scan ports, and more.

## Install

```bash
brew tap lu-zhengda/tap
brew install netwhiz
```

## Quick Start

```bash
netwhiz           # Launch interactive TUI
netwhiz --help    # Show all commands
```

## Commands

| Command | Description                                   |
|---------|-----------------------------------------------|
| `info`  | Show network overview (IP, gateway, DNS)      |
| `dns`   | DNS diagnostics and lookups                   |
| `wifi`  | WiFi network information                      |
| `ping`  | Ping a host                                   |
| `trace` | Trace route to a host                         |
| `speed` | Network speed test                            |
| `scan`  | Network port/device scan                      |

## TUI

Launch without arguments for interactive mode. Run diagnostics, view results, and navigate between tools with a keyboard-driven interface.

<!-- Screenshot placeholder: ![netwhiz TUI](docs/tui.png) -->

## License

MIT
