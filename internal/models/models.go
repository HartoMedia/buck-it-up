// internal/models/models.go

package models

// models.go

type Bucket struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

type AccessKeyRole string

const (
	RoleReadOnly  AccessKeyRole = "readOnly"
	RoleReadWrite AccessKeyRole = "readWrite"
	RoleAll       AccessKeyRole = "all"
)

type AccessKey struct {
	ID         int64         `json:"id"`
	BucketID   int64         `json:"bucket_id"`
	KeyID      string        `json:"key_id"`
	SecretHash string        `json:"-"`
	Role       AccessKeyRole `json:"role"`
	CreatedAt  int64         `json:"created_at"`
}

type Object struct {
	ID          int64  `json:"id"`
	BucketID    int64  `json:"bucket_id"`
	ObjectKey   string `json:"object_key"`
	FilePath    string `json:"file_path"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Checksum    string `json:"checksum"`
	CreatedAt   int64  `json:"created_at"`
}
