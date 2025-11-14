# Quick Reference - Authentication System

## Authentication Header Format
```
Authorization: Bearer <key_id>:<secret>
```

## Access Level Matrix

| Endpoint | No Auth | Read-Only | Read-Write | All |
|----------|---------|-----------|------------|-----|
| `GET /health` | ✓ | ✓ | ✓ | ✓ |
| `GET /echo` | ✓ | ✓ | ✓ | ✓ |
| `LIST /` | ✓* | ✓* | ✓* | ✓* |
| `POST /` | ✓* | ✓* | ✓* | ✓* |
| `GET /{name}` | ✗ | ✓ | ✓ | ✓ |
| `LIST /{bucketName}` | ✗ | ✓ | ✓ | ✓ |
| `GET /{bucketName}/all/*` | ✗ | ✓ | ✓ | ✓ |
| `GET /{bucketName}/metadata/*` | ✗ | ✓ | ✓ | ✓ |
| `GET /{bucketName}/content/*` | ✗ | ✓ | ✓ | ✓ |
| `POST /{bucketName}/upload` | ✗ | ✗ | ✓ | ✓ |
| `DELETE /{bucketName}/*` | ✗ | ✗ | ✓ | ✓ |
| `DELETE /{bucketName}` | ✗ | ✗ | ✗ | ✓ |

*Currently unprotected, will require admin auth in future

## Usage Examples

### Create Bucket
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"name":"myBucket"}'
```

### Get Bucket Info (Read-Only)
```bash
curl -H "Authorization: Bearer KEY_ID:SECRET" \
     http://localhost:8080/myBucket
```

### Upload Object (Read-Write)
```bash
curl -X POST http://localhost:8080/myBucket/upload \
  -H "Authorization: Bearer KEY_ID:SECRET" \
  -H "Content-Type: application/json" \
  -d '{"object_key":"file.txt","content":"data","content_type":"text/plain"}'
```

### Delete Bucket (All Access)
```bash
curl -X DELETE http://localhost:8080/myBucket \
  -H "Authorization: Bearer KEY_ID:SECRET"
```

## Error Responses

| Code | Meaning |
|------|---------|
| 401 | Missing/invalid credentials |
| 403 | Insufficient permissions or wrong bucket |
| 404 | Bucket/object not found |
| 500 | Internal server error |

## Important Notes

⚠️ **Secrets are shown only once** during bucket creation. Store them securely!

🔒 **Each access key is tied to ONE bucket only.** Cannot access other buckets.

📊 **Three keys per bucket**: Read-Only, Read-Write, All Access

🛡️ **Secrets are hashed** using SHA-256. Cannot be recovered from database.

