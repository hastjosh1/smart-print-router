# Smart Print Router

A small Windows tool that makes one "virtual printer" automatically send your
**labels** and **reports** to the right physical printer — with **2-up label
printing** (two labels side by side) that the browser app can't do on its own.

- **Labels / barcodes** (small page) → label printer (e.g. TVS LP46), composed 2-up
- **Reports** (A4 / Letter) → report printer (e.g. Canon LBP2900B)
- **No changes to your browser app** — it just prints to "Smart Print Router"

It auto-detects label vs report by **page size** (with a job-name keyword
fallback), and label sizes are **configurable** via `config.json`.

## ⬇️ Download (no build needed)

Grab the ready-to-run zip from the **[Releases](../../releases)** page, extract
it on your Windows PC, and follow **[TESTING.md](TESTING.md)**.

The `.exe` files are prebuilt — you do **not** need to install Go to run it.

## How it works

```
Browser app  →  "Smart Print Router" (virtual printer)
   →  Windows spooler  →  generic PostScript driver
   →  Redmon (port redirector)  →  router.exe (this project)
        ├─ label  → 2-up compose → label printer
        └─ report →               report printer
```

This is **software**, not a hand-written printer driver: it reuses Windows' own
spooler + a generic driver + the open-source [Redmon](http://www.ghostgum.com.au/software/redmon.htm)
port redirector, and our Go program (`router.exe`) does the routing. Silent
printing uses [SumatraPDF](https://www.sumatrapdfreader.org/).

## Testing in 3 stages

1. **Logic only** (no install) — drag a PDF onto `1-test-dryrun.bat`, preview the 2-up sheet.
2. **Real printing** (SumatraPDF) — drag a PDF onto `2-test-print.bat`, it prints to the right printer.
3. **Virtual printer** (Redmon) — run `install\install.ps1`, then print to "Smart Print Router".

Full walkthrough: **[TESTING.md](TESTING.md)**. Setup details: [install/README.md](install/README.md).

## Configuration

Edit `config.json` — printer names, label-size detection, and `label_profiles`
(add a new sticker type by appending an entry). See [install/README.md](install/README.md).

## Build from source

```sh
go test ./config/... ./detector/... ./cmd/router/...   # run tests
./build-dist.sh                                         # produce dist/ + zip
```

## Project layout

```
cmd/router/     per-job router (reads PDF on stdin, routes)
cmd/tray/       background tray app (status)
config/         settings + label-profile matching
detector/       page-size detection
composer/       2-up composition (pdfcpu)
printer/        silent printing (SumatraPDF)
install/        Windows install/uninstall scripts + setup guide
windows/        drag-and-drop test helpers (.bat)
```

See [CLAUDE.md](CLAUDE.md) for architecture notes and [FINDINGS.md](FINDINGS.md)
for analysis of a real barcode PDF.
