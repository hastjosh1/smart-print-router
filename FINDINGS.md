# Findings — overnight work on the real barcode PDF

You shared `document.pdf` (the barcode from your browser software). I inspected
it, fixed the code against the real pdfcpu API, verified the 2-up composition
end-to-end, and added tests. Summary below.

## What your barcode PDF actually is

| Property | Value |
|---|---|
| Page size | **113.04 × 56.16 pt = ~39.9 × 19.8 mm** (not 50 × 25 mm) |
| Page count | **2 pages — identical labels** (sample patient barcode; details redacted) |
| Barcode | **raster image, 276 × 120 px**, DeviceRGB (not vector) |
| Generator | react-pdf |

Two takeaways:

1. **The artwork is ~40 × 20 mm, but you said the physical sticker is 50 × 25 mm.**
   Same 2:1 shape, but the PDF is smaller than the label. See "Open question" below.
2. **The PDF already contains the 2 copies** you want side by side — so the 2-up
   step just places page 1 and page 2 next to each other. No duplication needed.

## What I verified (actually ran, not theory)

- Installed Go was present; built the router. The build **caught real pdfcpu API
  mismatches** in my first draft (wrong `NUpFile`/`PDFGridConfig` signatures) —
  now fixed against pdfcpu v0.8.1.
- Discovered pdfcpu's `grid` mode treats `dimensions` as the **per-cell** size and
  grows the sheet to (rows × cols). Fixed the composer accordingly.
- Ran the router on your real PDF in a new **`-dryrun`** mode and rendered the
  output: **two barcodes side by side on an ~80 × 20 mm sheet, at native
  resolution.** Confirmed visually. ✅
- Added unit tests for label-vs-report detection, profile matching, and job-name
  routing. **All passing** (`go test ./config/... ./detector/... ./cmd/router/...`).

## New `-dryrun` flag (for testing without a printer)

```sh
go build -o router ./cmd/router
./router -in document.pdf -config config.json -dryrun -out result.pdf
```

Writes the routed/composed PDF to `-out` instead of printing, and logs the
decision. Great for previewing 2-up layout on any machine.

## ⚠️ Open question that needs your input — the size mismatch

The barcode prints at **40 × 20 mm**, your sticker is **50 × 25 mm**. Three ways
to handle it, in order of preference:

1. **Print at native 40 × 20, centered on the 50 × 25 label** (whitespace border).
   Best barcode quality — no upscaling of the low-res raster. Recommended.
2. **Scale up to fill 50 × 25.** Fills the sticker but the barcode is only
   ~175 DPI native, so scaling drops it to ~140 DPI — below the 203 DPI thermal
   head. Risk: softer bars, less reliable scanning.
3. **Fix it at the source** — set your browser software to generate 50 × 25 mm
   labels (ideally higher-res barcodes). Cleanest long-term, but needs a change
   in that app.

**I also still need to know your label stock:** is the TVS LP46 loaded with
**2-across** die-cut labels (two 50 × 25 side by side per row) or a **single**
50 × 25 roll? If single-roll, true side-by-side 2-up isn't physical — we'd
instead print the 2 pages as 2 sequential labels (set `columns: 1` in config).
The code already supports both; I just need to set the right mode.

## Current config (config.json)

The label profile now matches the **real** ~40 × 20 mm size so 2-up triggers:

```json
"label_profiles": [
  { "name": "Barcode label (2-up side-by-side)",
    "width_mm": 39.9, "height_mm": 19.8, "tolerance_mm": 4,
    "two_up": { "enabled": true, "columns": 2, "gap_mm": 2 } }
]
```

## Suggested next steps when you're back

1. Tell me: **2-across stock or single roll?** and **native vs. fill** (option 1/2/3).
2. Confirm exact Windows printer names (`Get-Printer`).
3. I'll finalize the composer sizing for your real stock, then we test on the
   actual TVS printer.
