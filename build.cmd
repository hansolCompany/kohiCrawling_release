@echo off
setlocal
chcp 65001 >nul 2>&1
powershell.exe -NoProfile -NoLogo -ExecutionPolicy Bypass -File "%~dp0build.ps1" %*
endlocal
