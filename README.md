# gokapi-tui

A terminal UI client for [Gokapi](https://github.com/Forceu/Gokapi), the self-hosted file sharing server.

## Features

- Browse all uploaded files in a scrollable list with name, size, expiry, downloads remaining, upload date, and password indicator
- Upload files via an interactive file picker
- Configure expiry days, download limit, and password per upload
- Copy a file's share link to the clipboard
- Delete files with a confirmation prompt

## Installation

```
go install github.com/yeniklas/gokapi-tui@latest
```

Or build from source:

```
git clone https://github.com/yeniklas/gokapi-tui
cd gokapi-tui
make build
```

## Configuration

Create `~/.config/gokapi-tui/config.yaml`:

```yaml
server_url: "https://your.gokapi.instance"
api_key: "your_api_key"
defaults:
  allowed_downloads: 1
  expiry_days: 7
  password: ""
```

The API key can be generated in the Gokapi web UI under the **API** menu.

## Usage

```
gokapi-tui [flags]
```

| Flag | Description |
|------|-------------|
| `--version` | Print the current version and exit |
| `--self-update` | Update to the latest release from GitHub |

### Key bindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` / `G` | Jump to top / bottom |
| `u` | Upload a file |
| `y` | Copy share link to clipboard |
| `d` | Delete selected file |
| `r` | Refresh file list |
| `q` / `ctrl+c` | Quit |

During upload, a form lets you override the default expiry and download settings before confirming.

## License

GPL-3.0
