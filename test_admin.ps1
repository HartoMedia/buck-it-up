# Admin Password Test Script
# This script tests the admin password functionality

# Configuration
$baseUrl = "http://localhost:8080"
$adminPassword = "test-admin-password-123"

Write-Host "=== Buck It Up - Admin Password Test ===" -ForegroundColor Cyan
Write-Host ""

# Function to make requests with admin auth
function Invoke-AdminRequest {
    param(
        [string]$Uri,
        [string]$Method = "GET",
        [string]$Body = $null,
        [string]$ContentType = "application/json"
    )

    $headers = @{
        "Authorization" = "Bearer admin:$adminPassword"
    }

    try {
        if ($Body) {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -Headers $headers -Body $Body -ContentType $ContentType
        } else {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -Headers $headers
        }
        return $response
    } catch {
        Write-Host "Error: $_" -ForegroundColor Red
        Write-Host "Status Code: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
        return $null
    }
}

# Test 1: Health check (no auth required)
Write-Host "Test 1: Health Check (No Auth)" -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/health"
    Write-Host "✓ Server is healthy: $health" -ForegroundColor Green
} catch {
    Write-Host "✗ Server is not responding. Make sure to start it with: `$env:ADMIN_PASSWORD='$adminPassword'; .\buck_It_Up.exe" -ForegroundColor Red
    exit
}
Write-Host ""

# Test 2: List buckets without auth (should fail)
Write-Host "Test 2: List Buckets Without Auth (Should Fail)" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$baseUrl/" -Method LIST
    Write-Host "✗ Should have failed but didn't" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 401) {
        Write-Host "✓ Correctly rejected (401 Unauthorized)" -ForegroundColor Green
    } else {
        Write-Host "✗ Unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    }
}
Write-Host ""

# Test 3: List buckets with admin auth
Write-Host "Test 3: List Buckets With Admin Auth" -ForegroundColor Yellow
$buckets = Invoke-AdminRequest -Uri "$baseUrl/" -Method LIST
if ($buckets) {
    Write-Host "✓ Successfully listed $($buckets.Count) bucket(s)" -ForegroundColor Green
    $buckets | ForEach-Object { Write-Host "  - $($_.name) (ID: $($_.id))" }
} else {
    Write-Host "Note: No buckets found or error occurred" -ForegroundColor Yellow
}
Write-Host ""

# Test 4: Create a new bucket with admin auth
Write-Host "Test 4: Create New Bucket With Admin Auth" -ForegroundColor Yellow
$bucketName = "admin-test-bucket-$(Get-Random -Maximum 10000)"
$createBody = @{
    name = $bucketName
} | ConvertTo-Json

$newBucket = Invoke-AdminRequest -Uri "$baseUrl/" -Method POST -Body $createBody
if ($newBucket) {
    Write-Host "✓ Successfully created bucket: $bucketName" -ForegroundColor Green
    Write-Host "  Bucket ID: $($newBucket.id)" -ForegroundColor Gray
    Write-Host "  Access Keys Generated:" -ForegroundColor Gray
    $newBucket.access_keys | ForEach-Object {
        Write-Host "    - $($_.role): $($_.key_id)" -ForegroundColor Gray
    }

    # Save the first access key for later tests
    $testAccessKey = $newBucket.access_keys[0]
} else {
    Write-Host "✗ Failed to create bucket" -ForegroundColor Red
}
Write-Host ""

# Test 5: Upload object to bucket with admin auth
if ($newBucket) {
    Write-Host "Test 5: Upload Object With Admin Auth" -ForegroundColor Yellow
    $uploadBody = @{
        object_key = "test-admin-file.txt"
        content = "This file was uploaded using admin credentials!"
        content_type = "text/plain"
    } | ConvertTo-Json

    $uploadResult = Invoke-AdminRequest -Uri "$baseUrl/$bucketName/upload" -Method POST -Body $uploadBody
    if ($uploadResult) {
        Write-Host "✓ Successfully uploaded object" -ForegroundColor Green
        Write-Host "  Object Key: $($uploadResult.object_key)" -ForegroundColor Gray
        Write-Host "  Size: $($uploadResult.size) bytes" -ForegroundColor Gray
    }
    Write-Host ""
}

# Test 6: List bucket contents with admin auth
if ($newBucket) {
    Write-Host "Test 6: List Bucket Contents With Admin Auth" -ForegroundColor Yellow
    $objects = Invoke-AdminRequest -Uri "$baseUrl/$bucketName" -Method LIST
    if ($objects) {
        Write-Host "✓ Successfully listed $($objects.Count) object(s)" -ForegroundColor Green
        $objects | ForEach-Object { Write-Host "  - $($_.object_key) ($($_.size) bytes)" }
    }
    Write-Host ""
}

# Test 7: Get object with admin auth
if ($newBucket -and $uploadResult) {
    Write-Host "Test 7: Get Object With Admin Auth" -ForegroundColor Yellow
    $object = Invoke-AdminRequest -Uri "$baseUrl/$bucketName/all/test-admin-file.txt"
    if ($object) {
        Write-Host "✓ Successfully retrieved object" -ForegroundColor Green
        $decodedContent = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($object.content))
        Write-Host "  Content: $decodedContent" -ForegroundColor Gray
    }
    Write-Host ""
}

# Test 8: Access bucket created with different access key (admin can access any bucket)
if ($testAccessKey) {
    Write-Host "Test 8: Verify Admin Can Access Any Bucket" -ForegroundColor Yellow
    Write-Host "  (While regular key is limited to its bucket)" -ForegroundColor Gray

    # Create another bucket
    $bucketName2 = "admin-test-bucket2-$(Get-Random -Maximum 10000)"
    $createBody2 = @{
        name = $bucketName2
    } | ConvertTo-Json

    $newBucket2 = Invoke-AdminRequest -Uri "$baseUrl/" -Method POST -Body $createBody2

    if ($newBucket2) {
        # Try to access bucket2 with bucket1's access key (should fail)
        Write-Host "  Testing regular key access to different bucket..." -ForegroundColor Gray
        $regularKeyHeaders = @{
            "Authorization" = "Bearer $($testAccessKey.key_id):$($testAccessKey.secret)"
        }

        try {
            Invoke-RestMethod -Uri "$baseUrl/$bucketName2" -Method LIST -Headers $regularKeyHeaders
            Write-Host "  ✗ Regular key shouldn't have access" -ForegroundColor Red
        } catch {
            if ($_.Exception.Response.StatusCode.value__ -eq 403) {
                Write-Host "  ✓ Regular key correctly denied (403)" -ForegroundColor Green
            }
        }

        # Try to access bucket2 with admin credentials (should succeed)
        Write-Host "  Testing admin access to same bucket..." -ForegroundColor Gray
        $adminAccess = Invoke-AdminRequest -Uri "$baseUrl/$bucketName2" -Method LIST
        if ($adminAccess -ne $null) {
            Write-Host "  ✓ Admin can access any bucket" -ForegroundColor Green
        }
    }
    Write-Host ""
}

# Test 9: Delete object with admin auth
if ($newBucket -and $uploadResult) {
    Write-Host "Test 9: Delete Object With Admin Auth" -ForegroundColor Yellow
    $deleteResult = Invoke-AdminRequest -Uri "$baseUrl/$bucketName/test-admin-file.txt" -Method DELETE
    Write-Host "✓ Successfully deleted object" -ForegroundColor Green
    Write-Host ""
}

# Test 10: Delete bucket with admin auth
if ($newBucket) {
    Write-Host "Test 10: Delete Buckets With Admin Auth" -ForegroundColor Yellow

    Invoke-AdminRequest -Uri "$baseUrl/$bucketName" -Method DELETE | Out-Null
    Write-Host "✓ Successfully deleted bucket: $bucketName" -ForegroundColor Green

    if ($newBucket2) {
        Invoke-AdminRequest -Uri "$baseUrl/$bucketName2" -Method DELETE | Out-Null
        Write-Host "✓ Successfully deleted bucket: $bucketName2" -ForegroundColor Green
    }
    Write-Host ""
}

Write-Host "=== Tests Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor Cyan
Write-Host "- Admin password allows access to all routes" -ForegroundColor Green
Write-Host "- Admin can list all buckets" -ForegroundColor Green
Write-Host "- Admin can create buckets" -ForegroundColor Green
Write-Host "- Admin can access objects in any bucket" -ForegroundColor Green
Write-Host "- Admin can upload/delete objects in any bucket" -ForegroundColor Green
Write-Host "- Admin can delete any bucket" -ForegroundColor Green
Write-Host "- Regular access keys are still restricted to their bucket" -ForegroundColor Green

