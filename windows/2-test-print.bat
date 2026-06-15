@echo off
REM ============================================================
REM  STAGE 2 TEST - real printing, still no Redmon.
REM  Needs SumatraPDF installed (path set in config.json).
REM  Drag a PDF onto this file: it routes + prints to the
REM  correct real printer (TVS for labels, Canon for reports).
REM ============================================================

REM Jump to this script's own folder (drag-and-drop changes the working dir).
cd /d "%~dp0"

if "%~1"=="" (
  echo Drag a PDF file onto this .bat to print it.
  pause
  exit /b
)

if not exist "%~dp0router.exe" (
  echo ERROR: router.exe not found next to this script.
  echo Make sure this .bat is inside the extracted SmartPrintRouter folder.
  pause
  exit /b
)

echo Routing + printing:
echo   "%~1"
echo.

router.exe -in "%~1" -config "%~dp0config.json" -job "%~n1"
echo.
echo router.exe exit code: %errorlevel%
echo.
echo If nothing printed, check the log:
echo   C:\SmartPrintRouter\logs\router.log
pause
