<#
Verify the winnetflow ETW collector on the current Windows machine.

Purpose
    Run this on every target Windows version (10 / Server 2016 / 2019 / 2022)
    before declaring the collector production-ready. It executes the unit
    suite plus the real-ETW smoke/lifecycle/stress tests and checks that no
    session is left behind.

Prerequisites
    - Administrator shell (real-time ETW sessions require it)
    - Go toolchain matching the repository go.mod
    - The DataKit repository checkout (the script runs inside it)

Usage
    cd <datakit-checkout>
    pwsh internal/plugins/inputs/winnetflow/verify/verify-winnetflow.ps1

Record the "MATRIX" line together with the stress ratio printed by
TestETWStressTCP for the OS version matrix.
#>

$ErrorActionPreference = 'Stop'

function Step([string]$name) {
    Write-Host ""
    Write-Host "==> $name" -ForegroundColor Cyan
}

Step "Prerequisites"
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(
    [Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Not running elevated; some environments still allow real-time ETW sessions. The smoke/stress tests below are the definitive gate and will fail with a clear error if the session cannot be created."
}
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go toolchain not found."
}

$os = [System.Environment]::OSVersion.VersionString
$arch = $env:PROCESSOR_ARCHITECTURE
Write-Host "MATRIX: $os arch=$arch"
Write-Host ("Go: " + (go version))

$allPassed = $false
try {
Step "Unit tests"
go test ./internal/plugins/inputs/winnetflow
if ($LASTEXITCODE -ne 0) {
    Write-Error "Unit tests failed."
}

Step "ETW smoke, session lifecycle and stress (real ETW session, L4 TCPIP)"
$env:DK_WINNETFLOW_SMOKE = '1'
$env:DK_WINNETFLOW_STRESS = '1'
go test -v ./internal/plugins/inputs/winnetflow `
    -run "TestETWSessionLifecycle|TestETWSmokeCollectTCP|TestETWStressTCP|TestL4TDHPathWithRealEvents" `
    -timeout 300s
if ($LASTEXITCODE -ne 0) {
    Write-Error "L4 ETW verification failed."
}

Step "HTTP.sys smoke and session lifecycle (real ETW session, L7 httpflow)"
go test -v ./internal/plugins/inputs/winnetflow `
    -run "TestHTTPSessionLifecycle|TestETWSmokeCollectHTTP|TestHTTPTDHPathWithRealEvents" `
    -timeout 300s
if ($LASTEXITCODE -ne 0) {
    Write-Error "L7 HTTP ETW verification failed."
}

Step "Full Input.Run end-to-end (real ETW sessions, real traffic, feeder)"
go test -v ./internal/plugins/inputs/winnetflow `
    -run "TestInputRunEndToEnd|TestInputRunDefaultFiltering|TestInputRunFinalFlush" `
    -timeout 300s
if ($LASTEXITCODE -ne 0) {
    Write-Error "Input.Run end-to-end verification failed."
}

Step "60s soak under sustained traffic (event loss / goroutine leak gate)"
go test -v ./internal/plugins/inputs/winnetflow `
    -run "TestInputRunSoak" `
    -timeout 300s
if ($LASTEXITCODE -ne 0) {
    Write-Error "Soak verification failed."
}

Step "Race detector (optional, needs a C toolchain for cgo)"
# Real ETW tests already ran above. Keep the optional race pass bounded to the
# deterministic unit suite instead of repeating stress/soak under instrumentation.
Remove-Item Env:DK_WINNETFLOW_SMOKE -ErrorAction SilentlyContinue
Remove-Item Env:DK_WINNETFLOW_STRESS -ErrorAction SilentlyContinue
if (Get-Command gcc -ErrorAction SilentlyContinue) {
    go test -race ./internal/plugins/inputs/winnetflow -timeout 300s
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Race detector reported issues."
    }
} else {
    Write-Warning "gcc not found; skipping -race (CGO_ENABLED=1 requires a C toolchain)."
}

$allPassed = $true
} finally {
    Step "Session cleanup check"
    $left = logman query -ets 2>$null | Select-String "datakit-winnetflow|datakit-winnetflow-http"
    if ($left) {
        Write-Error "ETW session was left behind: $left"
    }
}

if ($allPassed) {
    Write-Host ""
    Write-Host "ALL CHECKS PASSED on $os ($arch)" -ForegroundColor Green
}
