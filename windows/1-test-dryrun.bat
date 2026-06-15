@echo off
REM ============================================================
REM  STAGE 1 TEST - no printer, no Redmon, no SumatraPDF needed.
REM  Drag a barcode/report PDF onto this file.
REM  It shows the routing decision and writes a *.routed.pdf you
REM  can open to preview the 2-up label layout.
REM ============================================================

REM Jump to this script's own folder (drag-and-drop changes the working dir).
cd /d "%~dp0"

if "%~1"=="" (
  echo Drag a PDF file onto this .bat to test it.
  pause
  exit /b
)

if not exist "%~dp0router.exe" (
  echo ERROR: router.exe not found next to this script.
  echo Make sure this .bat is inside the extracted SmartPrintRouter folder.
  pause
  exit /b
)

echo Running router on:
echo   "%~1"
echo.

router.exe -in "%~1" -config "%~dp0config.json" -dryrun -out "%~dpn1.routed.pdf"
echo.
echo router.exe exit code: %errorlevel%
echo.

if exist "%~dpn1.routed.pdf" (
  echo Result written next to your PDF:
  echo   "%~dpn1.routed.pdf"
  echo Opening it now...
  start "" "%~dpn1.routed.pdf"
) else (
  echo ERROR: no output was produced. Read any message above.
  echo Also check the log file: C:\SmartPrintRouter\logs\router.log
)

pause
