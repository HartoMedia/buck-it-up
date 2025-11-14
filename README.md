Quick start

This service is a small HTTP server that uses an embedded SQLite database (modernc.org/sqlite, no cgo required).

Build (Windows PowerShell):

```powershell
& "C:\Program Files\Go\bin\go.exe" build -o .\buck_It_Up.exe .
```

Run (PowerShell, set port and DB path if you want):

```powershell
$env:PORT = '8082'
$env:BUCK_DB_PATH = 'data.db'
.\buck_It_Up.exe
```

Smoke tests (PowerShell):

```powershell
# health
Invoke-RestMethod -Uri http://localhost:8082/health
# index
Invoke-RestMethod -Uri http://localhost:8082/
# echo
Invoke-RestMethod -Uri http://localhost:8082/echo -Method Post -Body 'hello'
# bucket lookup (should be 404 initially)
try { Invoke-RestMethod -Uri http://localhost:8082/buckets/test -ErrorAction Stop } catch { $_.Exception.Response.StatusCode }
```

Notes:
- The project uses modernc.org/sqlite so it doesn't require cgo (fixes the previous go-sqlite3 "CGO_ENABLED=0" issue).
- Default DB file is data.db in the repo root. It will be created automatically and migrations applied on first run.
- **Admin Access:** Set `ADMIN_PASSWORD` environment variable to enable admin access to all routes. See [ADMIN_ACCESS.md](ADMIN_ACCESS.md) for details.
- **Web UI:** Access the admin dashboard at `/ui` for a graphical interface. See [WEBUI_DOCUMENTATION.md](WEBUI_DOCUMENTATION.md) for details.

# Buck It Up

A simple object storage service with bucket management and access control.

## Features

- 🪣 **Bucket Management** - Create, list, and delete buckets
- 📦 **Object Storage** - Upload, download, and delete objects with metadata
- 🔐 **Access Control** - Role-based access keys (read-only, read-write, all)
- 👑 **Admin Access** - Global admin password for full system access
- 🌐 **Web UI** - Modern browser-based admin dashboard at `/ui`
- 💾 **SQLite Storage** - Embedded database with automatic migrations
- 🚀 **No CGO Required** - Pure Go implementation using modernc.org/sqlite

## Web UI

Buck It Up includes a built-in web interface for easy administration:

**Access:** Open `http://localhost:8080/ui` in your browser

**Features:**
- Visual bucket and object management
- Drag-and-drop file uploads
- Content preview for text files
- One-click downloads
- Mobile-responsive design

**Getting Started:**
```powershell
# Set admin password
$env:ADMIN_PASSWORD = "your-secure-password"

# Start server
.\buck_It_Up.exe

# Open browser to http://localhost:8080/ui
```

See [WEBUI_DOCUMENTATION.md](WEBUI_DOCUMENTATION.md) for complete UI documentation.

## Authentication

Buck It Up uses two authentication methods:

1. **Access Keys** - Bucket-specific credentials with role-based permissions (read-only, read-write, all)
2. **Admin Password** - Global administrator access to all buckets and routes (required for Web UI)

See [AUTHENTICATION.md](AUTHENTICATION.md) for access key details and [ADMIN_ACCESS.md](ADMIN_ACCESS.md) for admin password setup.

## Quick Admin Usage

```powershell
# Set admin password
$env:ADMIN_PASSWORD = "your-secure-password"

# Use admin credentials
$headers = @{
    "Authorization" = "Bearer admin:your-secure-password"
}

# List all buckets (admin only)
Invoke-RestMethod -Uri http://localhost:8080/ -Method LIST -Headers $headers

# Create a bucket (admin only)
$body = @{ name = "my-bucket" } | ConvertTo-Json
Invoke-RestMethod -Uri http://localhost:8080/ -Method POST -Headers $headers -Body $body -ContentType "application/json"
```

## Endpoints

### Get Object

GET /{bucketName}/{objectKey}

Returns JSON:
{
  "object": { /* metadata */ },
  "content": "raw file content"  // text; binary objects may need base64 later
}

Object file paths are stored under `data/buckets/<bucket_id>/objects/<object_id>`.
