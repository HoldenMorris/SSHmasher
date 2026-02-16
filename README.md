# SSHmasher

A local tool for managing your `~/.ssh` directory — keys, config, known_hosts, and backups — through a web GUI or native desktop app.

Built with Go, [Templ](https://templ.guide), [HTMX](https://htmx.org), and [Pico CSS](https://picocss.com). Desktop mode uses [Wails v2](https://wails.io).

## Features

- **Key Management** — List, generate (ed25519/RSA/ECDSA), inspect, edit comment, and delete SSH key pairs
- **Config Editor** — View and edit `~/.ssh/config` hosts via structured form or raw text editor, with duplicate detection
- **Known Hosts** — Browse, search, filter, and remove known_hosts entries
- **Backup & Restore** — Create tar.gz snapshots of `~/.ssh`, restore with automatic safety backup
- **Dark Mode** — Toggle between light, dark, and auto (system) themes

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl) |
| templ | latest | `go install github.com/a-h/templ/cmd/templ@latest` |
| ssh-keygen | any | Bundled with OpenSSH (used for key generation) |

Desktop mode has additional platform-specific requirements — see [Building the Desktop App](#building-the-desktop-app) below.

## Quick Start (Web Server)

```bash
# Install templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files and run
make dev
```

Open [http://127.0.0.1:8932](http://127.0.0.1:8932) in your browser.

### CLI Flags

```
-addr     Listen address (default: 127.0.0.1:8932)
-ssh-dir  SSH directory to manage (default: ~/.ssh)
```

Example: `go run ./cmd/server -addr 0.0.0.0:9000 -ssh-dir /tmp/test-ssh`

## Building

### Web Server (all platforms)

```bash
make build          # → bin/sshmasher
```

No CGO or platform-specific libraries needed. The binary embeds all static assets.

### Building the Desktop App

The desktop app wraps the same web UI in a native window using Wails v2. This requires CGO and platform-specific webview libraries.

#### Linux

**Dependencies:**

```bash
# Ubuntu/Debian 24.04+
sudo apt-get install -y build-essential libgtk-3-dev libwebkit2gtk-4.1-dev

# Ubuntu/Debian 22.04 and older
sudo apt-get install -y build-essential libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install -y gcc-c++ gtk3-devel webkit2gtk4.1-devel

# Arch
sudo pacman -S --needed gcc gtk3 webkit2gtk-4.1
```

**Build:**

```bash
# Ubuntu 24.04+ (webkit2gtk-4.1)
make build-desktop
# or explicitly:
CGO_ENABLED=1 go build -tags "dev,webkit2_41" -o bin/sshmasher-desktop ./cmd/desktop

# Ubuntu 22.04 and older (webkit2gtk-4.0)
DESKTOP_TAGS=dev make build-desktop
# or explicitly:
CGO_ENABLED=1 go build -tags "dev" -o bin/sshmasher-desktop ./cmd/desktop
```

For production builds, replace `dev` with `production`:

```bash
DESKTOP_TAGS=production,webkit2_41 make build-desktop
```

#### macOS

**Dependencies:**

Xcode command line tools (provides the required WebKit framework):

```bash
xcode-select --install
```

No additional libraries are needed — macOS includes WebKit natively.

**Build:**

```bash
CGO_ENABLED=1 go build -tags "dev" -o bin/sshmasher-desktop ./cmd/desktop

# Production build
CGO_ENABLED=1 go build -tags "production" -o bin/sshmasher-desktop ./cmd/desktop
```

#### Windows

**Dependencies:**

- A C compiler: install [MSYS2](https://www.msys2.org/) and then from the MSYS2 terminal:
  ```bash
  pacman -S mingw-w64-x86_64-gcc
  ```
- Add `C:\msys64\mingw64\bin` to your `PATH`
- WebView2 runtime (included in Windows 10 21H2+ and Windows 11; older systems download from [Microsoft](https://developer.microsoft.com/en-us/microsoft-edge/webview2/))

**Build (from PowerShell or cmd):**

```powershell
set CGO_ENABLED=1
go build -tags "dev" -o bin\sshmasher-desktop.exe .\cmd\desktop

# Production build (hides console window)
go build -tags "production" -ldflags "-H windowsgui" -o bin\sshmasher-desktop.exe .\cmd\desktop
```

### Cross-Compilation Notes

Cross-compiling the desktop app is difficult due to CGO dependencies. Build natively on each platform or use CI with platform-specific runners. The web server binary (`cmd/server`) cross-compiles trivially:

```bash
GOOS=linux   GOARCH=amd64 go build -o bin/sshmasher-linux   ./cmd/server
GOOS=darwin  GOARCH=arm64 go build -o bin/sshmasher-macos   ./cmd/server
GOOS=windows GOARCH=amd64 go build -o bin/sshmasher.exe     ./cmd/server
```

## Make Targets

| Target | Description |
|--------|-------------|
| `make dev` | Generate templ + run web server |
| `make dev-desktop` | Generate templ + run desktop app |
| `make build` | Build web server binary |
| `make build-desktop` | Build desktop binary |
| `make test` | Run all tests |
| `make generate` | Generate templ Go files |
| `make clean` | Remove binaries and generated files |

## Project Structure

```
SSHmasher/
├── cmd/
│   ├── server/main.go          # Web server entrypoint
│   └── desktop/main.go         # Wails desktop entrypoint
├── internal/
│   ├── model/types.go          # Domain types
│   ├── ssh/                    # Service layer (keys, config, known_hosts, backup)
│   ├── handler/                # HTTP handlers and router
│   └── view/                   # Templ components
├── static/                     # Vendored CSS/JS (Pico, HTMX)
├── docs/                       # Documentation, plans, TODOs
├── Makefile
└── wails.json
```

## API

All API endpoints return HTML partials when called with `HX-Request: true` (HTMX), or JSON otherwise.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/keys` | List all SSH keys |
| POST | `/api/keys` | Generate a new key |
| GET | `/api/keys/{name}` | Get key details |
| DELETE | `/api/keys/{name}` | Delete a key pair |
| GET | `/api/config/hosts` | List SSH config hosts |
| POST | `/api/config/hosts` | Add a host |
| GET | `/api/config/hosts/{alias}` | Get host details |
| PUT | `/api/config/hosts/{alias}` | Update a host |
| DELETE | `/api/config/hosts/{alias}` | Delete a host |
| GET | `/api/config/raw` | Get raw config text |
| PUT | `/api/config/raw` | Overwrite raw config |
| GET | `/api/knownhosts` | List known hosts |
| DELETE | `/api/knownhosts/{line}` | Remove entry by line |
| GET | `/api/knownhosts/raw` | Get raw known_hosts text |
| PUT | `/api/knownhosts/raw` | Overwrite raw known_hosts |
| GET | `/api/backup` | List backups |
| POST | `/api/backup` | Create backup |
| POST | `/api/backup/{filename}/restore` | Restore from backup |
| DELETE | `/api/backup/{filename}` | Delete a backup |

## License

TBD
