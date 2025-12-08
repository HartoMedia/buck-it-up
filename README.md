# Buck It Up  

[![Go](https://img.shields.io/badge/go-1.20+-00ADD8)](https://golang.org) [![License: MIT](https://img.shields.io/badge/license-MIT-brightgreen)](LICENSE)

Short Description

Buck It Up is a lightweight object storage HTTP service with bucket management, role-based access keys, an admin password mode, and a built-in web admin UI. It uses an embedded SQLite database and stores object files on disk so you can self-host small-to-medium workloads without external object stores.

Table of Contents

- Features
- Installation
- Usage Example
- Configuration
- API / Docs
- Contributing
- Roadmap

---
## Features

- Bucket management: create, list, and delete buckets
- Object storage: upload, download, preview, and delete objects
- Role-based access keys: readOnly, readWrite, and all scopes (per-bucket)
- Admin password: global admin access for full system control
- Built-in Web UI at /ui for visual administration
- SQLite-based persistence with automatic migrations
- No cgo required (uses modernc.org/sqlite)

---
## Build:

Docker (recommended)

1. Set all the necessary environment variables in the docker-compose.yml file (defaults are set).
2. Run the docker-compose command:
```bash
docker compose up -d
```

Everywhere else:
1. Set the environment variables (mainly the BUCKITUP_ADMIN_PASSWORD for accessing the UI).
2. Build the go binary for your setup
3. Run the binary

---
## Usage Example

Basic health check and endpoints:

```bash
# Health
curl http://localhost:8080/health
# Echo
curl -X GET http://localhost:8080/echo \
  -H "Content-Type: text/plain" \
  -d "Hello from curl!"
```

Access the UI at http://localhost:8080/ui


---
## Configuration


Key environment variables:
- PORT: HTTP port (default 8080 and only important if you don't use docker)
- BUCKITUP_DB_PATH: SQLite DB file path (default data.db)
- BUCKITUP_DATA_PATH: Root path for stored object files (default ./data)
- BUCKITUP_ADMIN_PASSWORD: Set to enable global admin access (required for /ui)
---
## API / Docs

- Authentication: Bearer tokens in the form `Authorization: Bearer <key_id>:<secret>` for access keys, or `Bearer admin:<BUCKITUP_ADMIN_PASSWORD>` for admin
- Bucket creation (admin only and doable in the ui): POST /
- List buckets (admin only): LIST /
- List bucket contents: LIST /{bucketName}
- Upload object: POST /{bucketName}/upload
- Get full object: GET /{bucketName}/all/{objectKey}
- Get metadata: GET /{bucketName}/metadata/{objectKey}
- Get content only: GET /{bucketName}/content/{objectKey}
- Delete object: DELETE /{bucketName}/{objectKey}


### Access Level Matrix

| Endpoint | No Auth | Read-Only | Read-Write | All |
|----------|---------|-----------|------------|-----|
| `GET /health` | ✓ | ✓ | ✓ | ✓ |
| `GET /echo` | ✓ | ✓ | ✓ | ✓ |
| `LIST /` | ✗ | ✗ | ✗ | ✓ |
| `POST /` | ✗ | ✗ | ✗ | ✓ |
| `GET /{name}` | ✗ | ✓ | ✓ | ✓ |
| `LIST /{bucketName}` | ✗ | ✓ | ✓ | ✓ |
| `GET /{bucketName}/all/*` | ✗ | ✓ | ✓ | ✓ |
| `GET /{bucketName}/metadata/*` | ✗ | ✓ | ✓ | ✓ |
| `GET /{bucketName}/content/*` | ✗ | ✓ | ✓ | ✓ |
| `POST /{bucketName}/upload` | ✗ | ✗ | ✓ | ✓ |
| `DELETE /{bucketName}/*` | ✗ | ✗ | ✓ | ✓ |
| `DELETE /{bucketName}` | ✗ | ✗ | ✗ | ✓ |

### yaak json

See [yaak.json](./yaak.json).

---
## Contributing

Contributions welcome! Suggested workflow:
1. Fork the repo
2. Create a feature branch
3. Open a pull request with a clear description

Please follow existing code style and write small, focused commits.

---
## Roadmap

- Web UI improvements
- Add CORS Settings for Buckets
- Improved large-file handling and streaming uploads
