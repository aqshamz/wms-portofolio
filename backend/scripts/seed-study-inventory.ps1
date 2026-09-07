# Development-only, additive inventory seed. It uses the Go inventory service;
# no SQL is executed from this script. Stable operation keys prevent duplicates.
[CmdletBinding()]
param(
    [string]$ActorUsername = ''
)

$ErrorActionPreference = 'Stop'
$backendPath = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$previousGoCache = $env:GOCACHE
Push-Location -LiteralPath $backendPath
try {
    $env:GOCACHE = Join-Path $env:TEMP 'wms-go-build-cache'
    $arguments = @('run', '-mod=readonly', './cmd/seed-study-inventory')
    if ($ActorUsername) {
        $arguments += @('-actor', $ActorUsername)
    }
    & go @arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Study inventory seed failed with exit code $LASTEXITCODE."
    }
} finally {
    $env:GOCACHE = $previousGoCache
    Pop-Location
}
