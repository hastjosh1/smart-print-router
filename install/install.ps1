<#
    Smart Print Router - installer
    --------------------------------
    Installs a virtual Windows printer ("Smart Print Router") whose output is
    piped to router.exe via Redmon, which auto-routes the PDF to the correct
    physical printer. Also installs the tray app and sets it to run at login.

    PREREQUISITES (install these first, links in install/README.md):
      1. Ghostscript  (PostScript->PDF printer driver path)
      2. Redmon 1.9   (port redirector)     -> http://www.ghostgum.com.au/software/redmon.htm
      3. SumatraPDF   (silent PDF printing)  -> https://www.sumatrapdfreader.org/
      4. router.exe + spr-tray.exe + config.json (built with `go build`, see README)

    Run from an elevated PowerShell:
        powershell -ExecutionPolicy Bypass -File install.ps1
#>

#Requires -RunAsAdministrator

param(
    [string]$PrinterName = "Smart Print Router",
    [string]$PortName    = "SPR:",
    [string]$InstallDir  = "C:\SmartPrintRouter"
)

$ErrorActionPreference = "Stop"

Write-Host "== Smart Print Router installer ==" -ForegroundColor Cyan

# 0. Show installed printers so the user can copy exact names into config.json.
Write-Host "`nInstalled printers on this PC:" -ForegroundColor Yellow
Get-Printer | Select-Object -ExpandProperty Name | ForEach-Object { Write-Host "   - $_" }
Write-Host "(Copy the Canon + TVS names exactly into config.json if they differ.)`n"

# 1. Lay down program files.
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $InstallDir "logs") | Out-Null

foreach ($f in @("router.exe", "spr-tray.exe", "config.json")) {
    $src = Join-Path $PSScriptRoot "..\$f"
    if (Test-Path $src) {
        Copy-Item $src $InstallDir -Force
    } else {
        Write-Host "WARNING: $f not found next to repo root - build it first." -ForegroundColor Red
    }
}

$routerPath = Join-Path $InstallDir "router.exe"
$trayPath   = Join-Path $InstallDir "spr-tray.exe"
Write-Host "Program installed to $InstallDir"

# 2. Configure the Redmon port to pipe job data into router.exe (stdin).
$redmonKey = "HKLM:\SYSTEM\CurrentControlSet\Control\Print\Monitors\Redirection Port\Ports\$PortName"
New-Item -Path $redmonKey -Force | Out-Null
Set-ItemProperty -Path $redmonKey -Name "Command"    -Value $routerPath
Set-ItemProperty -Path $redmonKey -Name "Arguments"  -Value "-config `"$InstallDir\config.json`""
Set-ItemProperty -Path $redmonKey -Name "Output"     -Value 2   # 2 = pipe job to program stdin
Set-ItemProperty -Path $redmonKey -Name "ShowWindow" -Value 0   # hidden
Set-ItemProperty -Path $redmonKey -Name "RunUser"    -Value 0
Write-Host "Redmon port '$PortName' configured."

# 3. Create the printer port (Redmon must already be installed as a monitor).
if (-not (Get-PrinterPort -Name $PortName -ErrorAction SilentlyContinue)) {
    Add-PrinterPort -Name $PortName -ErrorAction SilentlyContinue
}

# 4. Install the virtual printer with a PostScript driver.
$driver = "Microsoft PS Class Driver"
if (-not (Get-PrinterDriver -Name $driver -ErrorAction SilentlyContinue)) {
    Add-PrinterDriver -Name $driver
}
if (Get-Printer -Name $PrinterName -ErrorAction SilentlyContinue) {
    Write-Host "Printer '$PrinterName' already exists - removing first." -ForegroundColor Yellow
    Remove-Printer -Name $PrinterName
}
Add-Printer -Name $PrinterName -DriverName $driver -PortName $PortName
Write-Host "Printer '$PrinterName' installed." -ForegroundColor Green

# 5. Run the tray app at login (per-user; tray icons can't live in a service).
if (Test-Path $trayPath) {
    $runKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
    Set-ItemProperty -Path $runKey -Name "SmartPrintRouterTray" -Value "`"$trayPath`""
    Write-Host "Tray app set to start at login." -ForegroundColor Green
    Start-Process $trayPath   # launch now so it's available immediately
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Verify printer names in $InstallDir\config.json (see list above)."
Write-Host "  2. Print a test PDF to '$PrinterName'."
Write-Host "  3. Watch $InstallDir\logs\router.log (or the tray status)."
