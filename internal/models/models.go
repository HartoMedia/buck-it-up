package models

type Bucket struct {
	ID        int64  `json:"-"`
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
	ID         int64         `json:"-"`
	BucketID   int64         `json:"bucket_id"`
	KeyID      string        `json:"key_id"`
	SecretHash string        `json:"-"`
	Role       AccessKeyRole `json:"role"`
	CreatedAt  int64         `json:"created_at"`
}

type Object struct {
	ID          int64  `json:"-"`
	BucketID    int64  `json:"bucket_id"`
	ObjectKey   string `json:"object_key"`
	FilePath    string `json:"-"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Checksum    string `json:"checksum"`
	CreatedAt   int64  `json:"created_at"`
}

// Dtos that will be returned to the client

type BucketResponse struct {
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

func (b *Bucket) ToResponse() *BucketResponse {
	return &BucketResponse{
		Name:      b.Name,
		CreatedAt: b.CreatedAt,
	}
}

type ObjectResponse struct {
	BucketID    int64  `json:"bucket_id"`
	ObjectKey   string `json:"object_key"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Checksum    string `json:"checksum"`
	CreatedAt   int64  `json:"created_at"`
}

func (o *Object) ToResponse() *ObjectResponse {
	return &ObjectResponse{
		BucketID:    o.BucketID,
		ObjectKey:   o.ObjectKey,
		Size:        o.Size,
		ContentType: o.ContentType,
		Checksum:    o.Checksum,
		CreatedAt:   o.CreatedAt,
	}
}

type ObjectWithContentResponse struct {
	BucketID    int64  `json:"bucket_id"`
	ObjectKey   string `json:"object_key"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Checksum    string `json:"checksum"`
	CreatedAt   int64  `json:"created_at"`
	Content     string `json:"content"`
}

func (o *Object) ToResponseWithContent(content string) *ObjectWithContentResponse {
	return &ObjectWithContentResponse{
		BucketID:    o.BucketID,
		ObjectKey:   o.ObjectKey,
		Size:        o.Size,
		ContentType: o.ContentType,
		Checksum:    o.Checksum,
		CreatedAt:   o.CreatedAt,
		Content:     content,
	}
}

type AccessKeyResponse struct {
	BucketID  int64         `json:"bucket_id"`
	KeyID     string        `json:"key_id"`
	Role      AccessKeyRole `json:"role"`
	CreatedAt int64         `json:"created_at"`
}

func (ak *AccessKey) ToResponse() *AccessKeyResponse {
	return &AccessKeyResponse{
		BucketID:  ak.BucketID,
		KeyID:     ak.KeyID,
		Role:      ak.Role,
		CreatedAt: ak.CreatedAt,
	}
}

type AccessKeyWithSecretResponse struct {
	BucketID  int64         `json:"bucket_id"`
	KeyID     string        `json:"key_id"`
	Secret    string        `json:"secret"`
	Role      AccessKeyRole `json:"role"`
	CreatedAt int64         `json:"created_at"`
}

func (ak *AccessKey) ToResponseWithSecret(secret string) *AccessKeyWithSecretResponse {
	return &AccessKeyWithSecretResponse{
		BucketID:  ak.BucketID,
		KeyID:     ak.KeyID,
		Secret:    secret,
		Role:      ak.Role,
		CreatedAt: ak.CreatedAt,
	}
}
