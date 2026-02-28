# Linux Build Resources

This directory contains Linux-specific build resources.

## Required Files

### appicon.png
- Application icon in PNG format
- Size: 256x256 pixels or larger
- Will be used for the desktop entry

## Desktop Entry

Wails will generate a .desktop file automatically. The desktop entry will be installed to:
- `/usr/share/applications/` (system-wide install)
- `~/.local/share/applications/` (user install)

## Building

### Debian/Ubuntu (.deb)

```bash
# Build the application
wails build -platform linux/amd64

# Create .deb package (requires dpkg-deb)
mkdir -p debian/DEBIAN
mkdir -p debian/usr/bin
mkdir -p debian/usr/share/applications
mkdir -p debian/usr/share/icons/hicolor/256x256/apps
mkdir -p debian/usr/share/openclaw

# Copy files
cp build/bin/OpenClaw-Installer debian/usr/bin/openclaw-installer
cp build/appicon.png debian/usr/share/icons/hicolor/256x256/apps/openclaw.png
cp openclaw.desktop debian/usr/share/applications/

# Create control file
cat > debian/DEBIAN/control << EOF
Package: openclaw-installer
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Depends: libgtk-3-0, libwebkit2gtk-4.0-37
Maintainer: OpenClaw Team <team@openclaw.org>
Description: OpenClaw Installer
 Cross-platform AI assistant installer
EOF

# Build package
dpkg-deb --build debian openclaw-installer_1.0.0_amd64.deb
```

### Red Hat/CentOS/Fedora (.rpm)

```bash
# Build the application
wails build -platform linux/amd64

# Create RPM spec and build (requires rpmbuild)
# See openclaw.spec template
```

### AppImage

```bash
# Build the application
wails build -platform linux/amd64

# Create AppImage (requires appimagetool)
mkdir -p AppDir/usr/bin
mkdir -p AppDir/usr/share/applications
mkdir -p AppDir/usr/share/icons/hicolor/256x256/apps

cp build/bin/OpenClaw-Installer AppDir/usr/bin/openclaw-installer
cp build/appicon.png AppDir/usr/share/icons/hicolor/256x256/apps/openclaw.png
cp openclaw.desktop AppDir/usr/share/applications/
cp openclaw.desktop AppDir/

# Download and run appimagetool
wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage
chmod +x appimagetool-x86_64.AppImage
./appimagetool-x86_64.AppImage AppDir OpenClaw-Installer-x86_64.AppImage
```

## Cross-Compilation

From other platforms:

```bash
# Install cross-compilation toolchain
# macOS: brew install FiloSottile/musl-cross/musl-cross
# Or use Docker

# Build with Docker
docker run --rm -v $(pwd):/app -w /app \
  wailsapp/xgo:latest \
  wails build -platform linux/amd64
```

## Dependencies

The built application requires:
- libgtk-3-0
- libwebkit2gtk-4.0-37 (or newer)

These are typically pre-installed on most modern Linux distributions with a desktop environment.
