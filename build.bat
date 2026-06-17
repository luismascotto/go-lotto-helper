@echo off

cd /d %~dp0

set SVCNAME="lotofacil-checker"
set BUILDPATH="./cmd/lotofacil-checker"

echo Building %SVCNAME%...
go build -o %SVCNAME%.exe %BUILDPATH%

if errorlevel 1 (
    echo Failed to build %SVCNAME%
    pause
    exit /b 1
)

timeout 2