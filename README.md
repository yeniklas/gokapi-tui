# gokapi-tui

A terminal UI client for [Gokapi](https://github.com/Forceu/Gokapi), the self-hosted file sharing server.

## Features

- Two-tab interface: **Uploads** and **File Requests**
- Browse all uploaded files in a scrollable list with name, size, expiry, downloads remaining, upload date, and password indicator
- Upload files via an interactive file picker with configurable expiry, download limit, and password
- Copy a file's share link to the clipboard
- Delete files with a confirmation prompt
- List and create file requests (guest upload links)
- Copy the guest upload URL for a file request to the clipboard
- Delete file requests with a confirmation prompt
- View files uploaded via a specific file request

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

#### General

| Key | Action |
|-----|--------|
| `1` | Switch to Uploads tab |
| `2` | Switch to File Requests tab |
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` / `G` | Jump to top / bottom |
| `r` | Refresh list |
| `q` / `ctrl+c` | Quit |

#### Uploads tab

| Key | Action |
|-----|--------|
| `u` | Upload a file |
| `y` | Copy share link to clipboard |
| `d` | Delete selected file |

During upload, a form lets you override the default expiry and download settings before confirming.

#### File Requests tab

| Key | Action |
|-----|--------|
| `n` | Create a new file request |
| `y` | Copy guest upload link to clipboard |
| `d` | Delete selected file request |
| `enter` | View files uploaded via the selected request |
| `esc` | Go back |

## License

GPL-3.0
