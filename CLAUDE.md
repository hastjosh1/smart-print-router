# Smart Print Router — Project Context

> ## 🔧 CURRENT STATUS / HANDOFF (read this first)
>
> Working on the real hardware: Windows 10, **Canon LBP2900B** (reports),
> **SNBC TVSE LP46 Dlite BPLE** (labels, TSPL2). Browser app (react-pdf) emits a
> 2-page barcode PDF, each page ~39.9×19.8 mm; physical sticker is **4″×1″
> 2-across** (two 50.8×25.4 mm labels per row, with a gap).
>
> **What works:** detection, 2-up composition at 50.8×25.4 cells → 101.6×25.4 mm
> sheet (verified), config loading, report routing.
>
> **The label-printing journey:** SumatraPDF + the Windows driver kept
> **rotating/shrinking** the label no matter the orientation/scale settings. So
> labels now print via **native TSPL** (`rawlabel` package): Ghostscript
> rasterizes the composed PDF to 203-DPI mono, packed into a TSPL `BITMAP`, sent
> RAW to the printer (bypasses the driver → no rotation/scaling). Verified by
> decoding the TSPL back to an image = pixel-perfect.
>
> **What's left (calibrate on the real printer):**
> 1. Ensure Ghostscript is found (`label_raw.gs_path` in config.json, or on PATH).
> 2. Test-print and check: upside down? → flip `label_raw.direction` (0/1).
>    Too light/dark? → `label_raw.density` (0-15). Row spacing? → `label_raw.gap_mm`.
> 3. Confirm the barcode scans and centering is right; nudge `label_width_mm`/`gap_mm`.
> 4. Then Stage 3: install the virtual printer (Redmon) so the browser app prints to it.
>
> **How to test:** `router.exe -in <pdf> -config config.json -dryrun -out out.pdf`
> writes the composed PDF and (for labels) a `.prn` of the TSPL to inspect.
> Live: drag a PDF onto `2-test-print.bat`. Log: `C:\SmartPrintRouter\logs\router.log`.
> Build the Windows zip with `./build-dist.sh`; releases are published on GitHub.
> Do NOT commit the sample patient PDF (repo is public).

A **virtual Windows printer** that auto-routes PDFs from a browser app:
- **Labels/barcodes** (small page) → label printer, composed **2-up**
- **Reports** (A4/Letter) → report printer
- **Zero changes** to the browser app — it just prints to "Smart Print Router".

## Architecture

```
Browser App → "Smart Print Router" (virtual printer)
  → Windows Spooler → Redmon port redirector
  → smart-print-router.exe (stdin = PDF)
      ├─ detector: read page size (pdfcpu)
      ├─ small page → composer: 2-up → printer: silent print → LABEL printer
      └─ A4/Letter → printer: silent print → REPORT printer
```

Detection is **page-size first** (pdfcpu reads first-page dims in mm), with a
**job-name keyword** fallback ("barcode/label/sticker" vs "report/invoice/order")
that overrides size when present.

Silent printing uses **SumatraPDF** (`-print-to "<name>" -silent`).

## Layout

```
cmd/router/main.go   per-job entry: stdin → detect → match profile → route
cmd/tray/main.go     long-running tray app: status (tails log) + quick actions
cmd/tray/assets/     tray.ico (placeholder; replace with a real icon)
config/config.go     loads config.json; MatchProfile() picks a label profile
detector/detector.go page-size detection + IsLabel()
composer/composer.go N-up composition (configurable columns/gap); dup single labels
printer/printer.go   SumatraPDF silent print
config.json          user-editable settings + label_profiles
install/             install.ps1, uninstall.ps1, README.md (setup guide)
```

Two binaries: **router.exe** (short-lived, one per print job, invoked by Redmon)
and **spr-tray.exe** (per-user startup app for tray status; does no printing).

## Build & test

```sh
go mod tidy
go build -o router.exe ./cmd/router
go build -ldflags "-H windowsgui" -o spr-tray.exe ./cmd/tray
# preview routing/2-up without printing (writes result.pdf):
./router -in sample.pdf -config config.json -dryrun -out result.pdf
# unit tests (logic packages; skips tray which needs the Windows/Cocoa backend):
go test ./config/... ./detector/... ./cmd/router/...
```

## Decisions / notes

- **Redmon** chosen over a custom Port Monitor DLL (too complex). It's old but
  works on Win10/11. A **watched-folder** fallback is documented in install/README.md.
- Both barcode and report arrive as PDF, so size is the cleanest discriminator.
- pdfcpu is pure Go (no cgo) — easy cross-build.

## Deployment (confirmed 2026-06-14)

- OS: Windows 10
- Report printer: **Canon LBP2900B** (A4 laser)
- Label printer: **TVS LP46 Delite** (thermal)
- Label PDF size: **50 × 25 mm** (profile "Barcode 50x25", 2-up by default)
- Tray icon: yes (spr-tray.exe), routing runs silently per job
- Sticker types: configurable via `label_profiles` in config.json

Sample PDF analysis (see FINDINGS.md): real page = ~39.9 × 19.8 mm (NOT 50×25),
**2 identical pages**, barcode is a 276×120 raster (~175 DPI), react-pdf. 2-up
composition verified end-to-end via `-dryrun` (two barcodes side by side, ~80×20
mm sheet, native res). Build/tests/cross-compile all green.

Still open:
- [ ] Confirm exact Windows printer names via `Get-Printer`
- [ ] Is the TVS stock 2-across (so 2-up is physical) or a single-label roll?
- [ ] Native size vs. scale-to-fill 50×25 (raster is low-res — see FINDINGS.md)
- [ ] Replace placeholder cmd/tray/assets/tray.ico with a real icon
- [ ] Do NOT commit the sample PDF — it contains a patient name (privacy)

## References

- Redmon: http://www.ghostgum.com.au/software/redmon.htm
- Ghostscript: https://www.ghostscript.com/
- pdfcpu: https://github.com/pdfcpu/pdfcpu
- SumatraPDF silent print: `SumatraPDF.exe -print-to "Name" -silent file.pdf`
