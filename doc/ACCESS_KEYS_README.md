# Access Keys

When you create a bucket, the API automatically generates 3 access keys with different permission levels.

## Creating a Bucket

**Request:**
```bash
POST /
Content-Type: application/json

{
  "name": "my-bucket"
}
```

**Response:**
```json
{
  "id": 1,
  "name": "my-bucket",
  "created_at": 1763132203,
  "access_keys": [
    {
      "id": 1,
      "bucket_id": 1,
      "key_id": "Xy9mK2lQ3vRs8nT5pW7uJ4aB6cE=",
      "role": "readOnly",
      "created_at": 1763132203,
      "secret": "btI9a3vQsLlRd2J7vkfiWPACnU2EfbJacWNRFJFHqQk="
    },
    {
      "id": 2,
      "bucket_id": 1,
      "key_id": "PA4dnkzT4FbSjxCuvpTec4IqX2M=",
      "role": "readWrite",
      "created_at": 1763132203,
      "secret": "iSMrH_OB928LBP3ElEe8bLosKApbaT36Iu4Q0zs8Ewg="
    },
    {
      "id": 3,
      "bucket_id": 1,
      "key_id": "5SzNY21kmIX2xxxG9ja52JEaSSI=",
      "role": "all",
      "created_at": 1763132203,
      "secret": "18Q17fPmfjBYINh8XyAW2wlm8OQ01u2Q33Y66AQoyo0="
    }
  ]
}
```

## Access Key Roles

Three access keys are generated for each bucket:

1. **readOnly** - Read-only access to bucket objects
2. **readWrite** - Read and write access to bucket objects
3. **all** - Full access including bucket management

## Important Security Notes

- **Store the secrets securely!** They are only returned once during bucket creation.
- The `secret` field is the plaintext secret that should be used for authentication.
- The API stores a SHA256 hash of the secret internally (the `secret_hash` field, which is not returned in responses).
- Each access key has a unique `key_id` that identifies it.

## Access Key Structure

- `id` - Internal database ID
- `bucket_id` - ID of the associated bucket
- `key_id` - Public identifier for the access key
- `role` - Permission level (readOnly, readWrite, or all)
- `created_at` - Unix timestamp of creation
- `secret` - The secret key (only returned during creation)

