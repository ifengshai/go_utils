@echo off
setlocal
chcp 65001 >nul
cd /d "%~dp0"
set "PATH=%~dp0lib;%PATH%"
"%~dp0example_opencv.exe" %*
endlocal
