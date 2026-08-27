<#
Generate HTTP.sys traffic (via .NET HttpListener) for the winnetflow httpflow
smoke test. Prints "TRAFFIC_OK requests=<n>" on success.

Usage
    powershell -NoProfile -ExecutionPolicy Bypass -File gen-httpsvc-traffic.ps1 -Port 18530
#>
param(
    [int]$Port = 18530,
    [int]$Requests = 20,
    [int]$PostTrafficDelayMS = 5000
)

$ErrorActionPreference = 'Stop'
$prefix = "http://127.0.0.1:$Port/"
$ready = Join-Path $env:TEMP "httpsvc_srv_ready_$Port.txt"
Remove-Item $ready -ErrorAction SilentlyContinue

$srvScript = @"
`$ErrorActionPreference = 'Stop'
`$listener = New-Object System.Net.HttpListener
`$listener.Prefixes.Add('$prefix')
`$listener.Start()
Set-Content -Path '$ready' -Value 'ready'
while (`$true) {
  try {
    `$ctx = `$listener.GetContext()
    if (`$ctx.Request.HttpMethod -eq 'HEAD') {
      `$ctx.Response.StatusCode = 200
      `$ctx.Response.Close()
    } else {
      `$body = [System.Text.Encoding]::UTF8.GetBytes('{""ok"":true}')
      `$ctx.Response.StatusCode = 200
      `$ctx.Response.ContentType = 'application/json'
      `$ctx.Response.ContentLength64 = `$body.Length
      `$ctx.Response.OutputStream.Write(`$body, 0, `$body.Length)
      `$ctx.Response.Close()
    }
  } catch { break }
}
"@

$srv = Start-Process powershell -ArgumentList '-NoProfile', '-ExecutionPolicy', 'Bypass', '-Command', $srvScript -WindowStyle Hidden -PassThru
try {
    $deadline = (Get-Date).AddSeconds(15)
    while (-not (Test-Path $ready) -and (Get-Date) -lt $deadline) {
        Start-Sleep -Milliseconds 200
    }
    if (-not (Test-Path $ready)) {
        throw "HTTP listener did not start in time"
    }

    1..$Requests | ForEach-Object {
        Invoke-WebRequest -Uri "http://127.0.0.1:$Port/path$($_)?a=1" -Method Get -UseBasicParsing -TimeoutSec 10 | Out-Null
    }
    Invoke-WebRequest -Uri "http://127.0.0.1:$Port/api/upload" -Method Post -Body 'data-payload' -ContentType 'text/plain' -UseBasicParsing -TimeoutSec 10 | Out-Null
    Invoke-WebRequest -Uri "http://127.0.0.1:$Port/" -Method Head -UseBasicParsing -TimeoutSec 10 | Out-Null

    # Real-time ETW sessions flush buffered events periodically. Keep the
    # short-lived test server queryable through one flush window so the test
    # can validate process attribution rather than scheduler timing.
    if ($PostTrafficDelayMS -gt 0) {
        Start-Sleep -Milliseconds $PostTrafficDelayMS
    }

    Write-Output "TRAFFIC_OK requests=$Requests server_pid=$($srv.Id) client_pid=$PID"
} finally {
    Stop-Process -Id $srv.Id -Force -ErrorAction SilentlyContinue
    Remove-Item $ready -ErrorAction SilentlyContinue
}
