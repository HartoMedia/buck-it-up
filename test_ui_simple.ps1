# Simple Web UI Test
Write-Host "Buck It Up - Web UI Quick Test" -ForegroundColor Cyan
Write-Host "===============================" -ForegroundColor Cyan

$adminPassword = "test-admin-password-123"
$baseUrl = "http://localhost:8080"

Write-Host "`n1. Setting admin password..." -ForegroundColor Yellow
$env:ADMIN_PASSWORD = $adminPassword
Write-Host "   Password set!" -ForegroundColor Green

Write-Host "`n2. Starting server..." -ForegroundColor Yellow
$serverJob = Start-Job -ScriptBlock {
    param($path, $pwd)
    $env:ADMIN_PASSWORD = $pwd
    Set-Location $path
    & ".\buck_It_Up.exe"
} -ArgumentList $PWD, $adminPassword

Start-Sleep -Seconds 3

Write-Host "`n3. Testing endpoints..." -ForegroundColor Yellow
try {
    $health = Invoke-WebRequest -Uri "$baseUrl/health" -Method Get
    Write-Host "   [OK] Health check: $($health.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] Health check failed" -ForegroundColor Red
    Stop-Job $serverJob; Remove-Job $serverJob
    exit 1
}

try {
    $loginPage = Invoke-WebRequest -Uri "$baseUrl/ui/login" -Method Get
    Write-Host "   [OK] Login page: $($loginPage.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] Login page: $($_.Exception.Message)" -ForegroundColor Red
}

try {
    $dashboard = Invoke-WebRequest -Uri "$baseUrl/ui/dashboard" -Method Get
    Write-Host "   [OK] Dashboard page: $($dashboard.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] Dashboard page: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n4. Testing API authentication..." -ForegroundColor Yellow
$headers = @{ "Authorization" = "Bearer admin:$adminPassword" }
try {
    $buckets = Invoke-RestMethod -Uri "$baseUrl/" -Method "LIST" -Headers $headers
    Write-Host "   [OK] Admin auth works! Buckets: $($buckets.Count)" -ForegroundColor Green
} catch {
    Write-Host "   [FAIL] Admin auth failed" -ForegroundColor Red
}

Write-Host "`n===============================" -ForegroundColor Cyan
Write-Host "Web UI is ready!" -ForegroundColor Green
Write-Host "`nOpen in browser: $baseUrl/ui" -ForegroundColor White
Write-Host "Login password: $adminPassword" -ForegroundColor White
Write-Host "`nPress Ctrl+C to stop the server" -ForegroundColor Yellow

Start-Process "$baseUrl/ui"

# Keep server running
try {
    while ($true) {
        if ($serverJob.State -ne "Running") {
            Write-Host "`nServer stopped!" -ForegroundColor Red
            break
        }
        Start-Sleep -Seconds 2
    }
} finally {
    Stop-Job $serverJob -ErrorAction SilentlyContinue
    Remove-Job $serverJob -ErrorAction SilentlyContinue
    Write-Host "`nServer stopped." -ForegroundColor Green
}

