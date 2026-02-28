# OpenClaw Installer

Cross-platform installer for OpenClaw AI assistant with web-based configuration UI. Supports Windows, macOS, and Linux with offline USB deployment capability.

## Project Overview

OpenClaw Installer is a Go-based installation tool that provides:
- **Cross-platform support**: Windows (amd64/arm64), macOS (Intel/Apple Silicon), Linux (amd64/arm64)
- **Offline installation**: Deploy from USB without internet connection
- **Web configuration**: Built-in HTTP server (port 18080) for setup
- **IM adapters**: Pre-configured support for WeCom, DingTalk, and Feishu
- **GUI installer**: Wails-based desktop application for one-click installation
- **Auto-update**: Built-in update mechanism for seamless upgrades

## Quick Start

```bash
# Build all components
./scripts/build.sh all

# Build for specific platform
./scripts/build.sh windows
./scripts/build.sh macos
./scripts/build.sh linux

# Build Wails GUI installer
cd wails-installer && ./build.sh

# Create release packages
./scripts/create-release.sh
```

## Project Structure

```
├── adapters/           # IM adapter configurations (WeCom/DingTalk/Feishu)
├── build/              # Build artifacts (gitignored)
├── dist/               # Distribution packages (gitignored)
├── docs/               # Documentation
│   ├── architecture.md # System architecture
│   ├── USER_GUIDE.md   # End-user documentation (Chinese)
│   └── *.md           # Design docs, test plans, etc.
├── frontend/           # Web configuration UI (static files)
├── installer/          # Command-line installer (Go)
│   ├── main.go        # CLI entry point
│   ├── server.go      # HTTP server for web UI
│   ├── config.go      # Configuration management
│   └── *_test.go      # Unit tests
├── release/            # Release package templates
│   └── OpenClaw-v1.0.0/ # Platform-specific install scripts
├── scripts/            # Build and utility scripts
│   ├── build.sh       # Main build script
│   ├── build-*.sh     # Platform-specific builds
│   └── create-release.sh # Release packaging
├── updater/            # Auto-update system
├── usb-template/       # USB deployment template files
└── wails-installer/    # Wails-based GUI installer
    ├── main.go        # Wails entry point
    ├── frontend/      # GUI frontend (HTML/CSS/JS)
    └── internal/      # Installer logic modules
```

## Common Commands

### Development

```bash
# Run tests
cd installer && go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Format code
go fmt ./...

# Lint
golangci-lint run
```

### Building

```bash
# Build all platforms
./scripts/build.sh all

# Build specific platform with version
VERSION=1.1.0 ./scripts/build.sh macos

# Build Wails GUI
cd wails-installer && ./build.sh

# Cross-compile (requires Docker)
./scripts/build-linux.sh --docker
```

### Testing

```bash
# Run all tests
cd installer && go test -v ./...

# Run integration tests
go test -tags=integration -v ./...

# Test specific component
go test -v -run TestInstaller ./...
```

### Release

```bash
# Create release packages
./scripts/create-release.sh

# Create with version bump
./scripts/create-release.sh --version 1.1.0

# Clean and rebuild
./scripts/create-release.sh --clean --build
```

## Architecture

### Components

1. **Command-Line Installer** (`installer/`)
   - Go-based CLI tool
   - Embeds static web UI using `//go:embed`
   - HTTP server for configuration (port 18080)
   - Cross-compilation for all platforms

2. **Wails GUI Installer** (`wails-installer/`)
   - Desktop application using Wails v2
   - Native window frame, no console
   - 4-step wizard: Welcome → Mode → Adapters → Progress
   - Platform detection and auto-architecture selection

3. **Web Configuration UI** (`frontend/`)
   - Single-file HTML/CSS/JS application
   - No build step required
   - Served by embedded HTTP server
   - Configuration validation and preview

4. **Auto-Updater** (`updater/`)
   - Background update checking
   - Delta updates to minimize download
   - Rollback support on failure
   - Platform-specific update mechanisms

### Build System

- **Primary**: `scripts/build.sh` - Main orchestration
- **Platform-specific**: `scripts/build-{windows,macos,linux}.sh`
- **Wails**: `wails-installer/build.sh` - GUI application builds
- **Release**: `scripts/create-release.sh` - Package creation

### Configuration Flow

1. User runs installer (CLI or GUI)
2. Platform and architecture auto-detected
3. Installation mode selected (system/user/portable)
4. Adapter configurations customized
5. Files installed to target directory
6. PATH updated (system/user)
7. Shortcuts created (platform-specific)
8. Service registered (Linux/macOS optional)

## Development Workflow

### Adding a Feature

1. **Write tests first** - Follow TDD approach
2. **Implement feature** - Keep platform-specific code isolated
3. **Test locally** - Use `./scripts/build.sh` for quick iteration
4. **Update docs** - USER_GUIDE.md for user-facing changes
5. **Create release** - Verify with `./scripts/create-release.sh --build`

### Adding IM Adapter

1. Create adapter config in `adapters/<name>/adapter.json`
2. Add configuration template in `adapters/<name>/`
3. Update `installer/config.go` if custom logic needed
4. Add tests in `installer/config_test.go`
5. Update documentation

### Platform-Specific Changes

- Use `internal/platform/` for platform abstractions
- Keep platform code in separate files with build tags
- Test on target platform or use Docker for Linux builds

## Testing Strategy

### Unit Tests

- Location: `*_test.go` alongside source files
- Run: `go test ./...`
- Coverage target: 70%+

### Integration Tests

- Location: `integration_test.go`
- Run: `go test -tags=integration`
- Tests end-to-end installation flow

### Manual Testing

1. Build for target platform
2. Copy to test machine or VM
3. Test fresh install
4. Test upgrade path
5. Test uninstall
6. Verify service status (Linux/macOS)

### Platform Testing Matrix

| Platform | Arch | Fresh Install | Upgrade | Uninstall |
|----------|------|--------------|---------|-----------|
| Windows 11 | amd64 | ✓ | ✓ | ✓ |
| Windows 10 | amd64 | ✓ | ✓ | ✓ |
| macOS 14 | arm64 | ✓ | ✓ | ✓ |
| macOS 13 | amd64 | ✓ | ✓ | ✓ |
| Ubuntu 22.04 | amd64 | ✓ | ✓ | ✓ |
| Ubuntu 22.04 | arm64 | ✓ | ✓ | ✓ |

## Deployment

### USB Deployment

1. Run `./scripts/create-release.sh`
2. Copy `release/OpenClaw-v1.0.0/` to USB drive
3. USB structure:
   ```
   /OpenClaw/
   ├── windows/
   ├── macos/
   ├── linux/
   ├── shared/
   └── README.txt
   ```
4. User runs platform-specific installer script

### Network Deployment

1. Upload release packages to distribution server
2. Update version manifest for auto-updater
3. Notify users of new version

### Code Signing

**macOS** (see `docs/macos-signing.md`):
```bash
./scripts/sign-macos.sh --cert "Developer ID" \
  --input dist/openclaw-installer-darwin-amd64 \
  --output dist/openclaw-installer-darwin-amd64-signed
```

**Windows**: Use `signtool.exe` with code signing certificate

## Key Dependencies

- **Go 1.22+**: Primary language
- **Wails v2**: Desktop GUI framework
- **WebView2**: Windows web engine (runtime required)

## CI/CD

GitHub Actions workflows in `.github/workflows/`:
- Build on push/PR
- Cross-platform testing
- Automated release creation on tag

## Resources

- **User Guide**: `docs/USER_GUIDE.md` (Chinese)
- **Architecture**: `docs/architecture.md`
- **Build Guide**: `BUILD.md`
- **Test Plan**: `docs/TEST_PLAN.md`

## Notes

- All user-facing documentation is in Chinese
- Build scripts support `VERSION` environment variable
- `dist/` and `release/` contain large binaries - use `.gitignore`
- USB template files should be kept minimal for fast copy
