<#
    Smart Print Router - uninstaller
    Removes the virtual printer, port, and program files.
#>

#Requires -RunAsAdministrator

param(
    [string]$PrinterName = "Smart Print Router",
    [string]$PortName    = "SPR:",
    [string]$InstallDir  = "C:\SmartPrintRouter"
)

$ErrorActionPreference = "SilentlyContinue"

Write-Host "Removing printer '$PrinterName'..."
Remove-Printer -Name $PrinterName

Write-Host "Removing port '$PortName'..."
Remove-PrinterPort -Name $PortName

$redmonKey = "HKLM:\SYSTEM\CurrentControlSet\Control\Print\Monitors\Redirection Port\Ports\$PortName"
Remove-Item -Path $redmonKey -Recurse -Force

Write-Host "Removing tray startup entry and stopping tray..."
Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "SmartPrintRouterTray" -ErrorAction SilentlyContinue
Stop-Process -Name "spr-tray" -Force -ErrorAction SilentlyContinue

Write-Host "Removing program files at $InstallDir..."
Remove-Item -Path $InstallDir -Recurse -Force

Write-Host "Done." -ForegroundColor Green
