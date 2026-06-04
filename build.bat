@echo off

cd /d %~dp0

echo Building lotofacil-checker...
go build -o lotofacil-checker.exe ./cmd/lotofacil-checker

if errorlevel 1 (
    echo Failed to build lotofacil-checker.
    exit /b 1
)
