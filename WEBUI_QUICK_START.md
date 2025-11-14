# 🚀 Buck It Up - Quick Start with Web UI

## Start Server with Web UI

```powershell
# 1. Set admin password
$env:ADMIN_PASSWORD = "my-secure-password"

# 2. Start server
.\buck_It_Up.exe

# 3. Open browser
Start-Process "http://localhost:8080/ui"
```

## Web UI Quick Reference

| Action | Location | Steps |
|--------|----------|-------|
| **Login** | `/ui/login` | Enter admin password → Login |
| **View Buckets** | `/ui/dashboard` | Automatically shown after login |
| **Create Bucket** | Dashboard | Click "+ Create Bucket" → Enter name → Create |
| **View Objects** | Dashboard | Click "View Objects" on bucket card |
| **Upload File** | Bucket view | Click "+ Upload Object" → Select file → Upload |
| **Upload Text** | Bucket view | Click "+ Upload Object" → Choose "Text Content" → Paste/type → Upload |
| **View Object** | Bucket view | Click "View" button on object row |
| **Download** | Bucket view | Click "Download" button on object row |
| **Delete Object** | Bucket view | Click "Delete" button → Confirm |
| **Delete Bucket** | Dashboard | Click "Delete" button → Confirm (must be empty) |
| **Logout** | Any page | Click "Logout" in top-right corner |

## API Endpoints (for CLI/automation)

All endpoints require: `Authorization: Bearer admin:YOUR_PASSWORD`

### Buckets
```powershell
# List all buckets
Invoke-WebRequest -Uri "http://localhost:8080/" -Method LIST -Headers @{"Authorization"="Bearer admin:password"}

# Create bucket
Invoke-RestMethod -Uri "http://localhost:8080/" -Method POST -Headers @{"Authorization"="Bearer admin:password"; "Content-Type"="application/json"} -Body '{"name":"my-bucket"}'

# Delete bucket
Invoke-RestMethod -Uri "http://localhost:8080/my-bucket" -Method DELETE -Headers @{"Authorization"="Bearer admin:password"}
```

### Objects
```powershell
# List objects in bucket
Invoke-WebRequest -Uri "http://localhost:8080/my-bucket" -Method LIST -Headers @{"Authorization"="Bearer admin:password"}

# Upload object (text)
$body = @{object_key="file.txt"; content="Hello World"; content_type="text/plain"} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/my-bucket/upload" -Method POST -Headers @{"Authorization"="Bearer admin:password"; "Content-Type"="application/json"} -Body $body

# Get object
Invoke-RestMethod -Uri "http://localhost:8080/my-bucket/all/file.txt" -Headers @{"Authorization"="Bearer admin:password"}

# Delete object
Invoke-RestMethod -Uri "http://localhost:8080/my-bucket/file.txt" -Method DELETE -Headers @{"Authorization"="Bearer admin:password"}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_PASSWORD` | (none) | **Required** for web UI - Admin password for authentication |
| `PORT` | `8080` | Server port |
| `BUCK_DB_PATH` | `data.db` | SQLite database file path |
| `BUCK_DATA_PATH` | `data` | Object storage directory |

## URLs

- **Web UI**: http://localhost:8080/ui
- **Health Check**: http://localhost:8080/health
- **API Root**: http://localhost:8080/

## Common Tasks

### Upload a file via Web UI
1. Navigate to bucket view
2. Click "+ Upload Object"
3. Enter object key (e.g., `documents/report.pdf`)
4. Select file or paste text
5. Click "Upload"

### Download an object
1. Find object in bucket view table
2. Click "Download" button
3. File saves to browser's download folder

### Create multiple objects with folder structure
Use forward slashes in object keys:
- `images/logo.png`
- `images/photos/photo1.jpg`
- `docs/readme.txt`

### Preview text files
1. Click "View" on any object
2. Text files show preview
3. Binary files show size info

## Tips

💡 **Object Keys**: Use slashes for folder-like organization (`folder/subfolder/file.txt`)

💡 **Content Types**: Auto-detected for file uploads, specify for text uploads

💡 **Browser Compatibility**: Works in Chrome, Firefox, Safari, Edge

💡 **Mobile**: Fully responsive - works on phones and tablets

💡 **Session**: Login session persists in browser until you logout

💡 **Security**: Use HTTPS in production to encrypt admin password

## Troubleshooting

**"Can't access UI"**
→ Check server is running: `http://localhost:8080/health`

**"Invalid password"**
→ Verify `$env:ADMIN_PASSWORD` matches your login password

**"Server won't start"**
→ Port 8080 might be in use: `Get-Process | Where-Object {$_.ProcessName -eq "buck_It_Up"} | Stop-Process`

**"Upload fails"**
→ Check object key has no invalid characters, check file size

**"Can't delete bucket"**
→ Delete all objects first, then delete bucket

## Test the Installation

```powershell
# Run the test script
.\test_ui_simple.ps1

# Or manually test:
$env:ADMIN_PASSWORD = "test123"
.\buck_It_Up.exe

# Then open: http://localhost:8080/ui
```

## Documentation Files

- `README.md` - Project overview
- `WEBUI_DOCUMENTATION.md` - Complete UI guide
- `WEBUI_IMPLEMENTATION.md` - Implementation details
- `ADMIN_ACCESS.md` - Admin authentication guide
- `AUTHENTICATION.md` - Access key guide
- This file: `WEBUI_QUICK_START.md` - Quick reference

---

**Enjoy your new web UI! 🎉**

