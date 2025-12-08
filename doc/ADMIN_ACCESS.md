# Admin Access Documentation

## Overview

The Buck It Up server now supports an admin password that grants full access to all routes and all buckets. This is useful for administrative tasks and bypasses the bucket-specific access key restrictions.

## Configuration

### Setting the Admin Password

The admin password is configured via the `ADMIN_PASSWORD` environment variable. 
[
]()**Windows PowerShell:**
```powershell
$env:ADMIN_PASSWORD="your-secure-admin-password"
.\buck_It_Up.exe
```

**Linux/Mac:**
```bash
export ADMIN_PASSWORD="your-secure-admin-password"
./buck_It_Up
```

**Docker:**
```bash
docker run -e ADMIN_PASSWORD="your-secure-admin-password" -p 8080:8080 buck_it_up
```

### Security Considerations

- **Never use a weak password** - Use a strong, randomly generated password
- **Keep the password secret** - Do not commit it to version control
- **Use HTTPS in production** - Admin credentials should always be transmitted over HTTPS
- If `ADMIN_PASSWORD` is not set, admin authentication will be disabled

## Usage

### Authentication Format

Admin authentication uses the same Bearer token format as regular access keys:

```
Authorization: Bearer admin:<admin_password>
```

Where:
- `admin` is the fixed username for admin access
- `<admin_password>` is the password set in the `ADMIN_PASSWORD` environment variable

### Example Requests

#### List All Buckets (Admin Only)

**PowerShell:**
```powershell
$adminPassword = "your-secure-admin-password"
$headers = @{
    "Authorization" = "Bearer admin:$adminPassword"
}

Invoke-WebRequest -Uri "http://localhost:8080/" -Method LIST -Headers $headers
```

**curl:**
```bash
curl -X LIST http://localhost:8080/ \
  -H "Authorization: Bearer admin:your-secure-admin-password"
```

#### Create a New Bucket (Admin Only)

**PowerShell:**
```powershell
$adminPassword = "your-secure-admin-password"
$headers = @{
    "Authorization" = "Bearer admin:$adminPassword"
    "Content-Type" = "application/json"
}
$body = @{
    name = "my-new-bucket"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/" -Method Post -Headers $headers -Body $body
```

**curl:**
```bash
curl -X POST http://localhost:8080/ \
  -H "Authorization: Bearer admin:your-secure-admin-password" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-new-bucket"}'
```

#### Access Any Bucket's Objects

Admin credentials can also be used to access any bucket's objects, bypassing the bucket-specific access key restrictions:

**PowerShell:**
```powershell
$adminPassword = "your-secure-admin-password"
$headers = @{
    "Authorization" = "Bearer admin:$adminPassword"
}

# List objects in any bucket
Invoke-WebRequest -Uri "http://localhost:8080/some-bucket" -Method LIST -Headers $headers

# Get object from any bucket
Invoke-RestMethod -Uri "http://localhost:8080/some-bucket/all/my-object.txt" -Headers $headers

# Upload to any bucket
$uploadBody = @{
    object_key = "new-file.txt"
    content = "Hello World"
    content_type = "text/plain"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/some-bucket/upload" -Method Post -Headers $headers -Body $uploadBody -ContentType "application/json"

# Delete from any bucket
Invoke-RestMethod -Uri "http://localhost:8080/some-bucket/my-object.txt" -Method Delete -Headers $headers
```

## Routes Accessible with Admin Password

With admin credentials, you have full access to:

### Bucket Management (Admin Required)
- `LIST /` - List all buckets
- `POST /` - Create a new bucket
- `DELETE /{name}` - Delete a bucket by name

### Bucket Information (Read-Only or Higher)
- `GET /{name}` - Get bucket details

### Object Operations (Read-Only or Higher)
- `LIST /{bucketName}` - List bucket contents
- `GET /{bucketName}/all/*` - Get object with metadata
- `GET /{bucketName}/metadata/*` - Get object metadata only
- `GET /{bucketName}/content/*` - Get object content only

### Object Modification (Read-Write or Higher)
- `POST /{bucketName}/upload` - Upload object
- `DELETE /{bucketName}/*` - Delete object

## Comparison: Admin vs Regular Access Keys

| Feature | Regular Access Key | Admin Password |
|---------|-------------------|----------------|
| Scope | Single bucket only | All buckets |
| Creation | Auto-generated per bucket | Manually configured |
| Permissions | Role-based (read-only, read-write, all) | Full access (all) |
| Bucket Management | Only owns bucket | All buckets |
| Use Case | Normal operations | Administration & management |

## Best Practices

1. **Use regular access keys for normal operations** - Only use admin credentials when necessary
2. **Rotate admin password regularly** - Change the password periodically for security
3. **Audit admin access** - Monitor logs for admin credential usage
4. **Limit admin password distribution** - Only share with trusted administrators
5. **Use different passwords per environment** - Dev, staging, and production should have different admin passwords

## Troubleshooting

### "admin authentication not configured"
The `ADMIN_PASSWORD` environment variable is not set. Set it before starting the server.

### "invalid admin password"
The password provided does not match the `ADMIN_PASSWORD` environment variable. Verify you're using the correct password.

### "authorization required"
The `Authorization` header is missing. Include it with the format `Bearer admin:<password>`.

### "invalid authorization format"
The authorization header format is incorrect. Use: `Authorization: Bearer admin:<password>`

