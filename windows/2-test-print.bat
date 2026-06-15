@echo off
REM ============================================================
REM  STAGE 2 TEST - real printing, still no Redmon.
REM  Needs SumatraPDF installed (path set in config.json).
REM  Drag a PDF onto this file: it routes + prints to the
REM  correct real printer (TVS for labels, Canon for reports).
REM ============================================================

if "%~1"=="" (
  echo Drag a PDF file onto this .bat to print it.
  pause
  exit /b
)

router.exe -in "%~1" -config config.json -job "%~n1"

echo.
echo Done. If nothing printed, check the log:
echo   C:\SmartPrintRouter\logs\router.log
pause
