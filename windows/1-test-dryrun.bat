@echo off
REM ============================================================
REM  STAGE 1 TEST - no printer, no Redmon, no SumatraPDF needed.
REM  Drag a barcode/report PDF onto this file.
REM  It shows the routing decision and writes a *.routed.pdf you
REM  can open to preview the 2-up label layout.
REM ============================================================

if "%~1"=="" (
  echo Drag a PDF file onto this .bat to test it.
  pause
  exit /b
)

router.exe -in "%~1" -config config.json -dryrun -out "%~dpn1.routed.pdf"

echo.
echo Result written next to your PDF: "%~dpn1.routed.pdf"
echo Opening it now...
start "" "%~dpn1.routed.pdf"
pause
