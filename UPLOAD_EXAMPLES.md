# Upload Object Examples

## Endpoint: POST /{bucketName}/upload

Upload an object to a bucket. The object will be stored in the database and written to the filesystem.

### Request Body Format

```json
{
  "object_key": "path/to/your/file.txt",
  "content": "Your file content here",
  "content_type": "text/plain"
}
```

**Fields:**
- `object_key` (required): The key/path for the object (can contain slashes for nested paths)
- `content` (required): The content of the file as a string
- `content_type` (optional): MIME type of the content. Defaults to `application/octet-stream`

### Response

**Status 201 Created** - Returns the created object metadata:

```json
{
  "id": 9,
  "bucket_id": 1,
  "object_key": "my-folder/hello.txt",
  "file_path": "data/buckets/1/objects/9",
  "size": 12,
  "content_type": "text/plain",
  "checksum": "12",
  "created_at": 1731590400
}
```

### Example: Upload a Text File

**PowerShell:**
```powershell
$body = @{
    object_key = "documents/readme.txt"
    content = "Hello, Buck It Up!"
    content_type = "text/plain"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/photos/upload" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

**Curl:**
```bash
curl -X POST http://localhost:8080/photos/upload \
  -H "Content-Type: application/json" \
  -d '{
    "object_key": "documents/readme.txt",
    "content": "Hello, Buck It Up!",
    "content_type": "text/plain"
  }'
```

### Example: Upload a JSON File

**PowerShell:**
```powershell
$jsonContent = @{
    name = "John Doe"
    email = "john@example.com"
} | ConvertTo-Json

$body = @{
    object_key = "users/john.json"
    content = $jsonContent
    content_type = "application/json"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/documents/upload" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

### Example: Upload HTML Content

**PowerShell:**
```powershell
$htmlContent = @"
<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body><h1>Hello World</h1></body>
</html>
"@

$body = @{
    object_key = "web/index.html"
    content = $htmlContent
    content_type = "text/html"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/projects/upload" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

### Example: Upload Code File

**PowerShell:**
```powershell
$codeContent = @"
package main

import "fmt"

func main() {
    fmt.Println("Hello from Buck It Up!")
}
"@

$body = @{
    object_key = "src/hello.go"
    content = $codeContent
    content_type = "text/x-go"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/projects/upload" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"
```

## Error Responses

### 400 Bad Request
- Invalid bucket name
- Invalid object key
- Invalid JSON body
- Missing required fields

### 404 Not Found
- Bucket does not exist

### 409 Conflict
- Object with the same key already exists in the bucket

### 500 Internal Server Error
- Failed to create storage directory
- Failed to write object file
- Failed to update object metadata

## After Uploading

Once uploaded, you can retrieve the object using:

1. **Full object (metadata + content):**
   ```
   GET /{bucketName}/{objectKey}
   ```

2. **Metadata only:**
   ```
   GET /{bucketName}/metadata/{objectKey}
   ```

3. **Content only:**
   ```
   GET /{bucketName}/content/{objectKey}
   ```

## File Storage Location

Objects are stored in the filesystem at:
```
data/buckets/{bucket_id}/objects/{object_id}
```

The file path is also stored in the database for reference.

