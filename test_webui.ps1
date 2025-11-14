# Test script for Buck It Up Web UI
# This script tests the web UI functionality

Write-Host "Buck It Up - Web UI Test Script" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$adminPassword = "test-admin-password-123"
$baseUrl = "http://localhost:8080"

Write-Host "1. Setting admin password..." -ForegroundColor Yellow
$env:ADMIN_PASSWORD = $adminPassword
Write-Host "   Admin password set to: $adminPassword" -ForegroundColor Green
Write-Host ""

Write-Host "2. Starting Buck It Up server..." -ForegroundColor Yellow
Write-Host "   Starting server in background..." -ForegroundColor Gray

# Start the server in background
$serverJob = Start-Job -ScriptBlock {
    param($path, $adminPassword)
    $env:ADMIN_PASSWORD = $adminPassword
    Set-Location $path
    & ".\buck_It_Up.exe"
} -ArgumentList $PWD, $adminPassword

# Wait for server to start
Write-Host "   Waiting for server to start..." -ForegroundColor Gray
Start-Sleep -Seconds 3

# Check if server is running
try {
    $healthCheck = Invoke-WebRequest -Uri "$baseUrl/health" -Method Get -ErrorAction Stop
    if ($healthCheck.StatusCode -eq 200) {
        Write-Host "   Server is running!" -ForegroundColor Green
    }
} catch {
    Write-Host "   ERROR: Server failed to start" -ForegroundColor Red
    Stop-Job $serverJob
    Remove-Job $serverJob
    exit 1
}
Write-Host ""

Write-Host "3. Testing Web UI endpoints..." -ForegroundColor Yellow

# Test login page
try {
    $loginPage = Invoke-WebRequest -Uri "$baseUrl/ui/login" -Method Get -ErrorAction Stop
    if ($loginPage.StatusCode -eq 200 -and $loginPage.Content -like "*Buck It Up*") {
        Write-Host "   ✓ Login page accessible" -ForegroundColor Green
    }
} catch {
    Write-Host "   ✗ Login page failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test dashboard redirect
try {
    $dashboardPage = Invoke-WebRequest -Uri "$baseUrl/ui/dashboard" -Method Get -ErrorAction Stop
    if ($dashboardPage.StatusCode -eq 200 -and $dashboardPage.Content -like "*Buckets*") {
        Write-Host "   ✓ Dashboard page accessible" -ForegroundColor Green
    }
} catch {
    Write-Host "   ✗ Dashboard page failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test UI root redirect
try {
    $uiRoot = Invoke-WebRequest -Uri "$baseUrl/ui" -Method Get -MaximumRedirection 0 -ErrorAction SilentlyContinue
    if ($uiRoot.StatusCode -eq 302) {
        Write-Host "   ✓ UI root redirects to login" -ForegroundColor Green
    }
} catch {
    if ($_.Exception.Response.StatusCode -eq 302) {
        Write-Host "   ✓ UI root redirects to login" -ForegroundColor Green
    }
}

Write-Host ""

Write-Host "4. Testing API with admin credentials..." -ForegroundColor Yellow

# Create auth header
$authHeader = @{
    "Authorization" = "Bearer admin:$adminPassword"
}

# Test listing buckets
try {
    $listBuckets = Invoke-RestMethod -Uri "$baseUrl/" -Method "LIST" -Headers $authHeader -ErrorAction Stop
    Write-Host "   ✓ Admin authentication works" -ForegroundColor Green
    Write-Host "   Current buckets: $($listBuckets.Count)" -ForegroundColor Gray
} catch {
    Write-Host "   ✗ Admin authentication failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

Write-Host "5. Opening Web UI in browser..." -ForegroundColor Yellow
Write-Host "   Opening: $baseUrl/ui" -ForegroundColor Gray
Start-Process "$baseUrl/ui"
Write-Host "   ✓ Browser opened" -ForegroundColor Green
Write-Host ""

Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Web UI Test Complete!" -ForegroundColor Cyan
Write-Host ""
Write-Host "Server is running. You can now:" -ForegroundColor White
Write-Host "  1. Use the web UI at: $baseUrl/ui" -ForegroundColor White
Write-Host "  2. Login with password: $adminPassword" -ForegroundColor White
Write-Host ""
Write-Host "Press Ctrl+C to stop the server or close this window." -ForegroundColor Yellow
Write-Host ""

# Keep the server running and show logs
try {
    while ($true) {
        # Check if job is still running
        if ($serverJob.State -ne "Running") {
            Write-Host "Server stopped unexpectedly!" -ForegroundColor Red
            Receive-Job $serverJob
            break
        }

        # Show any output from the server
        $output = Receive-Job $serverJob -Keep
        if ($output) {
            Write-Host $output -ForegroundColor Gray
        }

        Start-Sleep -Seconds 1
    }
} finally {
    # Cleanup
    Write-Host ""
    Write-Host "Stopping server..." -ForegroundColor Yellow
    Stop-Job $serverJob -ErrorAction SilentlyContinue
    Remove-Job $serverJob -ErrorAction SilentlyContinue
    Write-Host "Server stopped." -ForegroundColor Green
}

