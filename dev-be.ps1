# watch-go.ps1
# Automatically rebuild and rerun a Go app when .go files change

$exe = "go-proxy.exe"
$hashFile = @{}

function Get-GoFileHashes {
    Get-ChildItem -Path . -Recurse -Filter *.go | ForEach-Object {
        $path = $_.FullName
        $lastWrite = $_.LastWriteTime.ToString("o")  # ISO 8601 format for consistent string comparison
        [PSCustomObject]@{
            Path = $path
            Hash = $lastWrite
        }
    }
}

function Start-GoApp {
    Write-Host "🔧 Building Go binary..."
    go build -o $exe $(ls *.go)
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Build failed. Waiting for next change..."
        return $null
    }

    Write-Host "🚀 Starting $exe..."
    Start-Process -FilePath ".\$exe" -PassThru -NoNewWindow
}

# Initial hash snapshot
$hashFile = Get-GoFileHashes | Group-Object -Property Path -AsHashTable -AsString
clear
$process = Start-GoApp

while ($true) {
    Start-Sleep -Seconds 1

    $newHashes = Get-GoFileHashes | Group-Object -Property Path -AsHashTable -AsString
    $changed = $false

    # Compare hash tables
    foreach ($path in $newHashes.Keys) {
        if (-not $hashFile.ContainsKey($path) -or $hashFile[$path].Hash -ne $newHashes[$path].Hash) {
            $changed = $true
            break
        }
    }

    # Detect deleted files too
    foreach ($path in $hashFile.Keys) {
        if (-not $newHashes.ContainsKey($path)) {
            $changed = $true
            break
        }
    }

    if ($changed) {
        clear
        Write-Host "`n⚡ Change detected, rebuilding..."
        if ($process -and !$process.HasExited) {
            Write-Host "🛑 Stopping old process..."
            Stop-Process -Id $process.Id -Force
        }
        $hashFile = $newHashes
        $process = Start-GoApp
    }
}
