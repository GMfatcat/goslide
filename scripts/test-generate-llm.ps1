# Manual LLM validation script for `goslide generate`.
# Runs two tests against OpenRouter:
#   1. Simple mode (topic: "Kubernetes")
#   2. Advanced mode (uses test-generate-llm-prompt.md alongside this script)
#
# Usage (PowerShell):
#   .\scripts\test-generate-llm.ps1
#
# Script is ASCII-only; Chinese content lives in the sibling .md file to
# avoid PowerShell 5.1 ANSI decoding issues.

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$testDir   = Join-Path $env:TEMP "goslide-llm-test"
$goslide   = "D:\CLAUDE-CODE-GOSLIDE\goslide.exe"
$srcPrompt = Join-Path $scriptDir "test-generate-llm-prompt.md"

if (-not (Test-Path $goslide)) {
    Write-Host "goslide.exe not found at $goslide. Run 'go build ./cmd/goslide' in the repo root first." -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $srcPrompt)) {
    Write-Host "prompt fixture not found at $srcPrompt" -ForegroundColor Red
    exit 1
}

# Prepare test directory
New-Item -ItemType Directory -Force -Path $testDir | Out-Null

# Write goslide.yaml (ASCII only)
$yamlLines = @(
    "generate:",
    "  base_url: https://openrouter.ai/api/v1",
    "  model: openai/gpt-oss-120b:free",
    "  api_key_env: OPENROUTER_API_KEY",
    "  timeout: 180s"
)
$yamlPath = Join-Path $testDir "goslide.yaml"
Set-Content -Path $yamlPath -Value ($yamlLines -join "`r`n") -Encoding UTF8 -NoNewline

# Copy the advanced-mode prompt file (contains Chinese)
Copy-Item -Path $srcPrompt -Destination (Join-Path $testDir "prompt.md") -Force

# Prompt for API key (not stored on disk or in shell history)
$secureKey = Read-Host -Prompt "Enter your OpenRouter API key" -AsSecureString
$bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secureKey)
$env:OPENROUTER_API_KEY = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
[System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr) | Out-Null

Push-Location $testDir
try {
    Write-Host ""
    Write-Host "=== Test 1: Simple mode (topic='Kubernetes') ===" -ForegroundColor Cyan
    & $goslide generate "Kubernetes" -o simple.md -f
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Test 1 failed (exit $LASTEXITCODE)" -ForegroundColor Red
    } else {
        Write-Host "Test 1 OK -> $testDir\simple.md" -ForegroundColor Green
    }

    Write-Host ""
    Write-Host "=== Test 2: Advanced mode (prompt.md, zh-TW, HS audience) ===" -ForegroundColor Cyan
    & $goslide generate prompt.md -o advanced.md -f
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Test 2 failed (exit $LASTEXITCODE)" -ForegroundColor Red
    } else {
        Write-Host "Test 2 OK -> $testDir\advanced.md" -ForegroundColor Green
    }
} finally {
    Pop-Location
    Remove-Item Env:OPENROUTER_API_KEY -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "Outputs:"
Write-Host "  $testDir\simple.md"
Write-Host "  $testDir\advanced.md"
Write-Host ""
Write-Host "(Artefact files *.raw.md / *.fixed.md written next to output on parse failure.)"
