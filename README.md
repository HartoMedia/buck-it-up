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

# Buck It Up

## Endpoints

### Get Object

GET /{bucketName}/{objectKey}

Returns JSON:
{
  "object": { /* metadata */ },
  "content": "raw file content"  // text; binary objects may need base64 later
}

Object file paths are stored under `data/buckets/<bucket_id>/objects/<object_id>`.
