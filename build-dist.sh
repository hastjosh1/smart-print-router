#!/usr/bin/env bash
# Build the Windows distribution folder + zip.
# Produces dist/SmartPrintRouter/ (ready to copy to a Windows PC) and
# dist/SmartPrintRouter.zip (attach to a GitHub Release).
#
# Requires Go. Run from the repo root:  ./build-dist.sh
set -euo pipefail

OUT="dist/SmartPrintRouter"
rm -rf dist
mkdir -p "$OUT"

echo "Building router.exe (windows/amd64)..."
GOOS=windows GOARCH=amd64 go build -o "$OUT/router.exe" ./cmd/router

echo "Building spr-tray.exe (windows/amd64, GUI)..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-H windowsgui" -o "$OUT/spr-tray.exe" ./cmd/tray

echo "Copying support files..."
cp config.json "$OUT/"
cp TESTING.md "$OUT/"
cp windows/*.bat "$OUT/"
cp -r install "$OUT/install"

echo "Zipping..."
( cd dist && zip -rq SmartPrintRouter.zip SmartPrintRouter )

echo "Done:"
ls -la dist
