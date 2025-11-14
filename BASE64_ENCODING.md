# Base64 Encoding in Buck It Up

## Overview

All object content is now base64-encoded when returned through the API, and the upload endpoint supports both raw and base64-encoded content.

## Retrieving Objects

### GET /{bucketName}/all/* (Metadata + Content)

Returns object metadata along with base64-encoded content:

```json
{
  "object": {
    "id": 1,
    "bucket_id": 1,
    "object_key": "myfile.txt",
    "file_path": "data/buckets/1/objects/1",
    "size": 1024,
    "content_type": "text/plain",
    "checksum": "1024",
    "created_at": 1234567890
  },
  "content": "SGVsbG8gV29ybGQh"  // base64-encoded
}
```

To decode the content:
```javascript
// JavaScript
const decoded = atob(response.content);

// Node.js
const decoded = Buffer.from(response.content, 'base64').toString();
```

```python
# Python
import base64
decoded = base64.b64decode(response['content'])
```

```go
// Go
import "encoding/base64"
decoded, err := base64.StdEncoding.DecodeString(response.Content)
```

```powershell
# PowerShell
$decoded = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($response.content))
```

### GET /{bucketName}/content/* (Content Only)

Returns raw binary content (not base64-encoded) with the appropriate Content-Type header.

```bash
curl -H "Authorization: Bearer key_id:secret" \
     http://localhost:8080/myBucket/content/myfile.txt
```

## Uploading Objects

### Option 1: Raw Text Content (Default)

Upload plain text content directly:

```bash
curl -X POST http://localhost:8080/myBucket/upload \
  -H "Authorization: Bearer key_id:secret" \
  -H "Content-Type: application/json" \
  -d '{
    "object_key": "test.txt",
    "content": "Hello World!",
    "content_type": "text/plain"
  }'
```

### Option 2: Base64-Encoded Content

Upload binary or encoded content using the `base64_encoded` flag:

```bash
curl -X POST http://localhost:8080/myBucket/upload \
  -H "Authorization: Bearer key_id:secret" \
  -H "Content-Type: application/json" \
  -d '{
    "object_key": "image.png",
    "content": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "content_type": "image/png",
    "base64_encoded": true
  }'
```

## Upload Request Schema

```json
{
  "object_key": "string (required)",      // The key/name of the object
  "content": "string (required)",          // The content (raw or base64)
  "content_type": "string (optional)",     // MIME type (default: application/octet-stream)
  "base64_encoded": "boolean (optional)"   // Set to true if content is base64 (default: false)
}
```

## Examples

### Upload and Retrieve a Text File

```powershell
# Upload
$uploadBody = @{
    object_key = "hello.txt"
    content = "Hello, World!"
    content_type = "text/plain"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/myBucket/upload" `
    -Method POST `
    -Headers @{ Authorization = "Bearer $keyId:$secret" } `
    -Body $uploadBody `
    -ContentType "application/json"

# Retrieve with metadata
$response = Invoke-RestMethod -Uri "http://localhost:8080/myBucket/all/hello.txt" `
    -Headers @{ Authorization = "Bearer $keyId:$secret" }

# Decode the base64 content
$decoded = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($response.content))
Write-Host $decoded  # Output: Hello, World!
```

### Upload a Binary File (Base64)

```powershell
# Read a file and encode it
$fileBytes = [System.IO.File]::ReadAllBytes("C:\path\to\image.png")
$base64Content = [System.Convert]::ToBase64String($fileBytes)

$uploadBody = @{
    object_key = "image.png"
    content = $base64Content
    content_type = "image/png"
    base64_encoded = $true
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/myBucket/upload" `
    -Method POST `
    -Headers @{ Authorization = "Bearer $keyId:$secret" } `
    -Body $uploadBody `
    -ContentType "application/json"
```

### Retrieve Binary Content (Raw)

```powershell
# Get raw content (not base64-encoded)
Invoke-WebRequest -Uri "http://localhost:8080/myBucket/content/image.png" `
    -Headers @{ Authorization = "Bearer $keyId:$secret" } `
    -OutFile "downloaded_image.png"
```

## Endpoint Summary

| Endpoint | Content Format | Use Case |
|----------|---------------|----------|
| `GET /{bucketName}/all/*` | JSON with base64 content | Get metadata + content programmatically |
| `GET /{bucketName}/metadata/*` | JSON (no content) | Get only metadata |
| `GET /{bucketName}/content/*` | Raw binary | Download file directly |
| `POST /{bucketName}/upload` | Raw text or base64 | Upload content |

## Why Base64?

- **JSON Compatibility**: Binary data can be safely transmitted in JSON
- **Text Safety**: Ensures binary content doesn't corrupt JSON structure
- **Cross-platform**: Works consistently across all platforms and languages
- **API Consistency**: Standard approach for binary data in REST APIs

## Notes

⚠️ **Base64 increases size by ~33%** - For large files, consider using the `/content/*` endpoint to get raw binary data

✅ **Both upload methods work** - Use raw text for simple strings, base64 for binary data

🔄 **Automatic handling** - The server automatically handles both formats based on the `base64_encoded` flag

