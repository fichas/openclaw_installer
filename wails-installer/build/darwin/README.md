# macOS Build Resources

This directory contains macOS-specific build resources.

## Required Files

### appicon.png
- Application icon in PNG format
- Size: 1024x1024 pixels (will be converted to .icns)
- Should have transparent background

## Generating Icons

```bash
# Create iconset from PNG
mkdir OpenClaw.iconset
sips -z 16 16     appicon.png --out OpenClaw.iconset/icon_16x16.png
sips -z 32 32     appicon.png --out OpenClaw.iconset/icon_16x16@2x.png
sips -z 32 32     appicon.png --out OpenClaw.iconset/icon_32x32.png
sips -z 64 64     appicon.png --out OpenClaw.iconset/icon_32x32@2x.png
sips -z 128 128   appicon.png --out OpenClaw.iconset/icon_128x128.png
sips -z 256 256   appicon.png --out OpenClaw.iconset/icon_128x128@2x.png
sips -z 256 256   appicon.png --out OpenClaw.iconset/icon_256x256.png
sips -z 512 512   appicon.png --out OpenClaw.iconset/icon_256x256@2x.png
sips -z 512 512   appicon.png --out OpenClaw.iconset/icon_512x512.png
sips -z 1024 1024 appicon.png --out OpenClaw.iconset/icon_512x512@2x.png

# Convert to .icns
iconutil -c icns OpenClaw.iconset
rm -rf OpenClaw.iconset
```

## Info.plist

Wails will generate Info.plist automatically, but you can customize it by providing your own.

## Building

```bash
# Build for macOS (universal binary)
wails build -platform darwin/universal

# Build for Intel Macs
wails build -platform darwin/amd64

# Build for Apple Silicon
wails build -platform darwin/arm64
```

## Code Signing (Optional)

For distribution outside the App Store:
```bash
# Sign the application
codesign --deep --force --verify --verbose --sign "Developer ID Application: Your Name" OpenClaw-Installer.app

# Create DMG
create-dmg \
  --volname "OpenClaw Installer" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --app-drop-link 450 185 \
  "OpenClaw-Installer.dmg" \
  "build/bin/OpenClaw-Installer.app"
```
