# Smart Print Router — Setup (Windows)

A virtual printer that auto-routes PDFs: **small page → label printer (2-up)**,
**A4/Letter → report printer**. Zero changes to your browser app — it just prints
to "Smart Print Router".

## 1. Prerequisites

Install these on the Windows machine first:

| Tool | Purpose | Link |
|---|---|---|
| **Go** (build only) | Compile the router | https://go.dev/dl/ |
| **Ghostscript** | PostScript → PDF path | https://www.ghostscript.com/releases/gsdnld.html |
| **Redmon 1.9** | Pipes the print job to our exe | http://www.ghostgum.com.au/software/redmon.htm |
| **SumatraPDF** | Silent printing to real printers | https://www.sumatrapdfreader.org/ |

> **Note on Redmon:** it's old (32-bit) but still works on Win10/11. Install the
> matching 32/64-bit build and register it as a print monitor per its README.
> If Redmon proves painful, see "Alternative: watched folder" below.

## 2. Build the router + tray

Two executables: `router.exe` (invoked per print job) and `spr-tray.exe`
(background tray app).

```sh
cd smart-print-router
go mod tidy
go build -o router.exe ./cmd/router
go build -ldflags "-H windowsgui" -o spr-tray.exe ./cmd/tray
```

(The `-H windowsgui` flag stops a console window flashing up for the tray.)

## 3. Configure

Edit `config.json` with your **exact** Windows printer names and any sticker
types you use. Defaults are pre-filled for this deployment:

```json
{
  "report_printer": "Canon LBP2900B",
  "label_printer": "TVS LP46 Delite",
  "sumatra_path": "C:/Program Files/SumatraPDF/SumatraPDF.exe",
  "log_file": "C:/SmartPrintRouter/logs/router.log",
  "label_detection": { "max_width_mm": 150, "max_height_mm": 200 },
  "default_label": { "two_up": { "enabled": false, "columns": 1, "gap_mm": 0 } },
  "label_profiles": [
    {
      "name": "Barcode 50x25",
      "width_mm": 50, "height_mm": 25, "tolerance_mm": 4,
      "two_up": { "enabled": true, "columns": 2, "gap_mm": 2 }
    }
  ]
}
```

**Adding a new sticker type:** append another object to `label_profiles` with
its `width_mm`/`height_mm` and `two_up` settings. The router matches an incoming
PDF's page size (orientation-independent, within `tolerance_mm`) to a profile and
uses its 2-up layout. Labels that match no profile use `default_label`.

Find exact printer names with (the installer also prints this list):

```powershell
Get-Printer | Select-Object Name
```

> Confirm the Canon shows as `Canon LBP2900B` (sometimes `Canon LBP2900`) and the
> TVS as `TVS LP46 Delite`, and copy the exact strings into config.json.

## 4. Install the virtual printer

From an **elevated** PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File install\install.ps1
```

This copies the exe + config to `C:\SmartPrintRouter`, wires the Redmon port to
pipe jobs into the exe, and creates the "Smart Print Router" printer.

## 5. Test

1. Print any A4 PDF to **Smart Print Router** → should land on the report printer.
2. Print a label-sized PDF → should compose 2-up and hit the label printer.
3. Watch the log: `C:\SmartPrintRouter\logs\router.log`.

The tray icon ("Smart Print Router") shows the latest log line and has menu
items to open the log, edit config.json, and quit. Routing keeps working even if
you quit the tray — the tray is status only.

You can also test the routing logic **without a printer** using dry-run mode,
which writes the routed/composed PDF to a file and logs the decision:

```sh
router.exe -in document.pdf -config config.json -dryrun -out result.pdf
# open result.pdf to preview the 2-up label sheet
```

Or test live routing (will actually print):

```sh
router.exe -in test-label.pdf -job "barcode-123"
router.exe -in test-report.pdf
```

## Uninstall

```powershell
powershell -ExecutionPolicy Bypass -File install\uninstall.ps1
```

## Alternative: watched folder (no driver pain)

If Redmon is too fiddly, skip the virtual printer entirely:

1. In the browser app's print dialog, use **"Microsoft Print to PDF"** (or
   "Save as PDF") and save into a watched folder, e.g. `C:\SmartPrintRouter\inbox`.
2. Run the router in `-watch` mode (TODO: add fsnotify watcher) so it picks up
   new PDFs from that folder and routes them.

This is far simpler to deploy but requires one extra click (choosing the save
folder) unless the browser app can be set to always save there.

## Troubleshooting

- **Nothing prints / no log file** → the exe isn't being invoked. Re-check the
  Redmon port `Command`/`Arguments`/`Output=2` registry values.
- **Wrong printer chosen** → tune `max_width_mm` / `max_height_mm` in config, or
  rename the print job to include "label"/"report" (job-name hint overrides size).
- **2-up looks wrong** → adjust `gap_mm`, or set `two_up.enabled = false` to print
  single labels while you debug.
