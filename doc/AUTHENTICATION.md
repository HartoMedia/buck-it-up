# Authentication System

## Overview

Buck It Up uses a key-based authentication system where each bucket has multiple access keys with different permission levels. Each access key is scoped to a single bucket and has a specific role that determines what operations it can perform.

## Access Levels

### No Authentication Required
The following endpoints are publicly accessible:
- `GET /health` - Health check endpoint
- `GET /echo` - Echo endpoint for testing

### Read-Only Access (`readOnly` role)
Access keys with the `readOnly` role can perform:
- All publicly accessible endpoints
- `GET /{bucketName}` - Get bucket information (only for their bucket)
- `LIST /{bucketName}` - List all objects in their bucket
- `GET /{bucketName}/all/*` - Get object metadata and content
- `GET /{bucketName}/metadata/*` - Get object metadata only
- `GET /{bucketName}/content/*` - Get object content only

### Read-Write Access (`readWrite` role)
Access keys with the `readWrite` role can perform:
- All read-only operations
- `POST /{bucketName}/upload` - Upload objects to their bucket
- `DELETE /{bucketName}/*` - Delete objects from their bucket

### All Access (`all` role)
Access keys with the `all` role can perform:
- All read-write operations
- `DELETE /{bucketName}` - Delete their bucket (including all objects and access keys)

## Authentication Format

All authenticated requests must include an `Authorization` header in the following format:

```
Authorization: Bearer <key_id>:<secret>
```

### Example
```bash
curl -H "Authorization: Bearer abc123:xyz789secret" \
     http://localhost:8080/myBucket
```

## How It Works

1. **Create a Bucket**: When you create a bucket via `POST /`, the system automatically generates three access keys:
   - One with `readOnly` role
   - One with `readWrite` role
   - One with `all` role

2. **Response**: The bucket creation response includes all three access keys with their secrets:
   ```json
   {
     "id": 1,
     "name": "myBucket",
     "created_at": 1234567890,
     "access_keys": [
       {
         "id": 1,
         "bucket_id": 1,
         "key_id": "readOnlyKeyId123",
         "role": "readOnly",
         "created_at": 1234567890,
         "secret": "readOnlySecret456"
       },
       {
         "id": 2,
         "bucket_id": 1,
         "key_id": "readWriteKeyId789",
         "role": "readWrite",
         "created_at": 1234567890,
         "secret": "readWriteSecret012"
       },
       {
         "id": 3,
         "bucket_id": 1,
         "key_id": "allAccessKeyId345",
         "role": "all",
         "created_at": 1234567890,
         "secret": "allAccessSecret678"
       }
     ]
   }
   ```

3. **Store Secrets Safely**: Save these secrets immediately! The secrets are only shown once during bucket creation. They are stored as SHA-256 hashes in the database and cannot be retrieved later.

4. **Use Access Keys**: Include the appropriate access key in your requests based on the operations you need to perform.

## Security Features

### Secret Hashing
- Secrets are hashed using SHA-256 before storage
- The database only stores the hash, not the plaintext secret
- Secrets are compared using constant-time comparison to prevent timing attacks

### Bucket Isolation
- Each access key is tied to a specific bucket
- Access keys cannot access objects or perform operations on other buckets
- The middleware validates bucket ownership on every request

### Role-Based Access Control
- Permissions are hierarchical (readOnly < readWrite < all)
- The middleware enforces minimum required permission levels
- Attempting operations without sufficient permissions returns `403 Forbidden`

## Error Responses

### 401 Unauthorized
- Missing `Authorization` header
- Invalid authorization format
- Invalid credentials (wrong key_id or secret)

### 403 Forbidden
- Insufficient permissions for the requested operation
- Attempting to access a bucket the key doesn't own

### 404 Not Found
- Bucket doesn't exist
- Object doesn't exist

## Admin Access

Bucket creation (`POST /`) and listing all buckets (`LIST /`) now require admin authentication. Admin access is protected by a password that can access all buckets and perform all administrative operations.

### Admin Authentication

To use admin operations, you must:
1. Set the `BUCKITUP_ADMIN_PASSWORD` environment variable
2. Use the admin credentials in your requests:
   ```
   Authorization: Bearer admin:<BUCKITUP_ADMIN_PASSWORD>
   ```

See [ADMIN_ACCESS.md](ADMIN_ACCESS.md) for complete admin documentation.

## Examples

### Creating a Bucket (Admin Only)
```bash
curl -X POST http://localhost:8080/ \
  -H "Authorization: Bearer admin:your_admin_password" \
  -H "Content-Type: application/json" \
  -d '{"name":"myBucket"}'
```

### Listing Bucket Contents (Read-Only Key)
```bash
curl -X LIST http://localhost:8080/myBucket \
  -H "Authorization: Bearer readOnlyKeyId123:readOnlySecret456"
```

### Uploading an Object (Read-Write Key)
```bash
curl -X POST http://localhost:8080/myBucket/upload \
  -H "Authorization: Bearer readWriteKeyId789:readWriteSecret012" \
  -H "Content-Type: application/json" \
  -d '{"object_key":"test.txt","content":"Hello World","content_type":"text/plain"}'
```

### Deleting a Bucket (All Access Key)
```bash
curl -X DELETE http://localhost:8080/myBucket \
  -H "Authorization: Bearer allAccessKeyId345:allAccessSecret678"
```

## Testing Authentication

You can test the authentication system by:

1. Creating a bucket and saving the access keys
2. Trying to access different endpoints with different role keys
3. Verifying that read-only keys cannot upload/delete
4. Verifying that keys cannot access other buckets
5. Testing with invalid credentials to ensure proper error handling

