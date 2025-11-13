// internal/models/objects.go

package models

// objects.go

import (
	"context"
	"database/sql"
)

type ObjectStore struct {
	db *sql.DB
}

func NewObjectStore(db *sql.DB) *ObjectStore {
	return &ObjectStore{db: db}
}

func (s *ObjectStore) PutObject(ctx context.Context, o *Object) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
        INSERT INTO objects (
            bucket_id, object_key, file_path, size, content_type, checksum, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `,
		o.BucketID, o.ObjectKey, o.FilePath, o.Size, o.ContentType, o.Checksum, o.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *ObjectStore) GetObject(ctx context.Context, bucketID int64, objectKey string) (*Object, error) {
	var o Object
	err := s.db.QueryRowContext(ctx, `
        SELECT id, bucket_id, object_key, file_path, size, content_type, checksum, created_at
        FROM objects
        WHERE bucket_id = ? AND object_key = ?
    `,
		bucketID, objectKey,
	).Scan(
		&o.ID, &o.BucketID, &o.ObjectKey, &o.FilePath, &o.Size,
		&o.ContentType, &o.Checksum, &o.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (s *ObjectStore) ListObjects(ctx context.Context, bucketID int64) ([]*Object, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, bucket_id, object_key, file_path, size, content_type, checksum, created_at
		FROM objects
		WHERE bucket_id = ?
	`, bucketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objects []*Object
	for rows.Next() {
		var o Object
		if err := rows.Scan(
			&o.ID, &o.BucketID, &o.ObjectKey, &o.FilePath, &o.Size,
			&o.ContentType, &o.Checksum, &o.CreatedAt,
		); err != nil {
			return nil, err
		}
		objects = append(objects, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return objects, nil
}

func (s *ObjectStore) DeleteObject(ctx context.Context, bucketID int64, objectKey string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM objects
		WHERE bucket_id = ? AND object_key = ?
	`,
		bucketID, objectKey,
	)
	return err
}

func (s *ObjectStore) ListObjectsByBucketName(ctx context.Context, bucketName string) ([]*Object, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT o.id, o.bucket_id, o.object_key, o.file_path, o.size, o.content_type, o.checksum, o.created_at
        FROM objects o
        JOIN buckets b ON b.id = o.bucket_id
        WHERE b.name = ?
        ORDER BY o.id
    `, bucketName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objects []*Object
	for rows.Next() {
		var o Object
		if err := rows.Scan(
			&o.ID, &o.BucketID, &o.ObjectKey, &o.FilePath, &o.Size,
			&o.ContentType, &o.Checksum, &o.CreatedAt,
		); err != nil {
			return nil, err
		}
		objects = append(objects, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return objects, nil
}
