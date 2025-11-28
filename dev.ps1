try {
    $be = $(Start-Process -FilePath wails -ArgumentList "dev" -PassThru -NoNewWindow)
    
    while ($true) {
        Start-Sleep -Seconds 30
        $(Get-Item frontend/vite.config.ts).LastWriteTime = (Get-Date)
    }
} finally {
    try { Stop-Process -Id $be.Id } catch {}
}