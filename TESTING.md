# Testing on your Windows 10 PC

Test in 3 stages — each one builds on the last, so if something breaks you know
exactly which layer it is. **You do NOT need Go installed** — the .exe files are
prebuilt and in this folder.

## What's in this folder

```
router.exe          the routing program
spr-tray.exe        the background tray app
config.json         your settings (printer names, label sizes)
1-test-dryrun.bat   Stage 1 helper (drag a PDF onto it)
2-test-print.bat    Stage 2 helper (drag a PDF onto it)
install\            Stage 3 scripts + full setup guide
```

Copy this whole `SmartPrintRouter` folder to your Windows PC (USB stick, network
share, OneDrive — whatever's easy). Put it somewhere simple like `C:\SmartPrintRouter`.

---

## Stage 1 — does the routing logic work? (2 minutes, nothing to install)

This proves size detection + 2-up composition with **no printer, no SumatraPDF,
no Redmon**.

1. Double-check `config.json` — the label profile and printer names.
2. **Drag your barcode PDF onto `1-test-dryrun.bat`.**
3. It opens a `*.routed.pdf` next to your file. For a barcode you should see
   **two barcodes side by side**. For an A4 report you'll see it unchanged.

✅ If the preview looks right, the brain works. Move to Stage 2.
❌ If wrong, send me the PDF + what you saw and I'll adjust.

> Tip: also drag a **report (A4) PDF** onto it to confirm reports are detected as
> reports (they should NOT be turned into 2-up).

---

## Labels print via native TSPL (needs Ghostscript)

By default the router prints **labels** by rendering them and sending the printer
its native **TSPL** commands directly — this bypasses the Windows driver, so the
label cannot be rotated or rescaled (which is what caused the earlier sideways /
shrunk prints). This needs **Ghostscript** installed:
https://www.ghostscript.com/releases/gsdnld.html (64-bit). The router auto-finds
it under `C:\Program Files\gs\...`.

Reports (A4 → Canon) still print normally via SumatraPDF. To revert labels to the
old driver-based printing, set `label_raw.enabled` to `false` in `config.json`.

Tuning lives in `config.json` under `label_raw`: `gap_mm` (gap between rows),
`direction` (flip 0/1 if upside down), `density` (darkness 0-15), `copies`.

---

## Stage 2 — does it print to the right real printer? (needs SumatraPDF + Ghostscript)

This proves silent printing to your actual Canon + TVS, **still without Redmon**.

1. Install **SumatraPDF**: https://www.sumatrapdfreader.org/download-free-pdf-viewer
   (the regular installer is fine). If you install it somewhere other than
   `C:\Program Files\SumatraPDF\SumatraPDF.exe`, update `sumatra_path` in config.json.
2. Confirm your exact printer names. Open **PowerShell** and run:
   ```powershell
   Get-Printer | Select-Object Name
   ```
   Make sure `report_printer` and `label_printer` in `config.json` match EXACTLY
   (copy-paste). The Canon is often `Canon LBP2900B` or `Canon LBP2900`.
3. **Drag your barcode PDF onto `2-test-print.bat`.** It should print to the TVS.
4. Drag a report PDF onto it — should print to the Canon.

✅ If both go to the correct printer, the whole engine works.
❌ If nothing prints, open `C:\SmartPrintRouter\logs\router.log` — it logs every
   decision and any error. Send me that file.

> At this stage you can already use it as a poor-man's router: just "print to PDF"
> from your browser app, then drag the PDF onto `2-test-print.bat`. Stage 3 removes
> that manual step.

---

## Stage 3 — the actual virtual printer (Redmon)

Only do this once Stages 1 & 2 work. This makes "Smart Print Router" appear as a
real printer so your browser app prints to it directly — no dragging files.

Full instructions are in **`install\README.md`**. Short version:

1. Install prerequisites: **Ghostscript** and **Redmon** (links in install\README.md).
2. Open **PowerShell as Administrator**, `cd` into this folder, and run:
   ```powershell
   powershell -ExecutionPolicy Bypass -File install\install.ps1
   ```
3. In your browser app's print dialog, choose **Smart Print Router**.
4. Watch `C:\SmartPrintRouter\logs\router.log` or the tray icon.

To remove everything: `powershell -ExecutionPolicy Bypass -File install\uninstall.ps1`

---

## If you get stuck

The log file is your friend: **`C:\SmartPrintRouter\logs\router.log`**. It records
the page size detected, which profile matched, and where it routed. Send me that
plus a screenshot and I can tell what happened.
