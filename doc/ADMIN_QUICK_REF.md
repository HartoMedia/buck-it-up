# Admin Password Quick Reference

## Setup
```powershell
# Set the admin password before starting the server
$env:BUCKITUP_ADMIN_PASSWORD = "your-secure-password"
.\buck_It_Up.exe
```

## Authentication Header
```
Authorization: Bearer admin:<your-password>
```

## PowerShell Helper
```powershell
# Store password in variable
$adminPassword = "your-secure-password"

# Create auth headers
$adminHeaders = @{
    "Authorization" = "Bearer admin:$adminPassword"
}

# Use in requests
Invoke-RestMethod -Uri "http://localhost:8080/" -Method LIST -Headers $adminHeaders
```

## Admin-Only Routes
- `LIST /` - List all buckets
- `POST /` - Create new bucket

## Admin Can Access All Routes
Admin credentials work for ANY route, including:
- All bucket operations
- All object operations  
- Access to any bucket (not restricted like regular keys)

## Test Script
Run the test script to verify admin functionality:
```powershell
.\test_admin.ps1
```

## Security Tips
✓ Use a strong, random password  
✓ Never commit the password to git  
✓ Use HTTPS in production  
✓ Use regular access keys for normal operations  
✓ Reserve admin for administrative tasks only  

## Example: Create & Access Bucket
```powershell
$adminPassword = "your-secure-password"
$headers = @{ "Authorization" = "Bearer admin:$adminPassword" }

# Create bucket
$bucket = Invoke-RestMethod -Uri "http://localhost:8080/" -Method POST -Headers $headers `
    -Body (@{ name = "new-bucket" } | ConvertTo-Json) -ContentType "application/json"

# Upload to bucket
$upload = Invoke-RestMethod -Uri "http://localhost:8080/new-bucket/upload" -Method POST -Headers $headers `
    -Body (@{ object_key = "file.txt"; content = "Hello"; content_type = "text/plain" } | ConvertTo-Json) `
    -ContentType "application/json"

# List bucket contents
Invoke-RestMethod -Uri "http://localhost:8080/new-bucket" -Method LIST -Headers $headers

# Get object
Invoke-RestMethod -Uri "http://localhost:8080/new-bucket/all/file.txt" -Headers $headers

# Delete object
Invoke-RestMethod -Uri "http://localhost:8080/new-bucket/file.txt" -Method DELETE -Headers $headers

# Delete bucket
Invoke-RestMethod -Uri "http://localhost:8080/new-bucket" -Method DELETE -Headers $headers
```

