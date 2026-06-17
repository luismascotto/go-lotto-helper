@echo off

cd /d %~dp0
set SVCNAME="lotofacil-checker"
call build.bat

if errorlevel 1 (
    exit /b 1
)

start "" "%SVCNAME%.exe" -config config-20260527.json
exit
