# Windows Build Resources

This directory contains Windows-specific build resources.

## Required Files

Before building for Windows, you need to place the following files here:

### icon.ico
- Application icon in ICO format
- Recommended sizes: 16x16, 32x32, 48x48, 256x256
- Can be generated from PNG using online tools or ImageMagick

### icon.rc (optional)
Resource file for embedding the icon. If not present, Wails will generate one automatically.

### manifest.xml (optional)
Windows application manifest for DPI awareness and other settings.

## Generating Icons

From a PNG image:
```bash
# Using ImageMagick
convert icon.png -define icon:auto-resize=256,128,64,48,32,16 icon.ico

# Or use online converters:
# https://convertio.co/png-ico/
# https://icoconvert.com/
```

## Building

```bash
# Build for Windows (from Linux/macOS with cross-compilation)
wails build -platform windows/amd64

# Build for Windows (native)
wails build -platform windows/amd64
```
