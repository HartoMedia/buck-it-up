// internal/models/objects.go

package models

// objects.go

import (
	"context"
	"database/sql"
)

type BucketStore struct {
	db *sql.DB
}

func NewBucketStore(db *sql.DB) *BucketStore {
	return &BucketStore{db: db}
}

func (s *BucketStore) NewBucket(ctx context.Context, b *Bucket) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
        INSERT INTO buckets (
            id, name, created_at
        ) VALUES (?, ?, ?)
    `,
		b.ID, b.Name, b.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *BucketStore) GetBucketByName(ctx context.Context, bucketName string) (*Bucket, error) {
	var b Bucket
	err := s.db.QueryRowContext(ctx, `
        SELECT id, name, created_at
        FROM buckets
        WHERE name = ?
    `,
		bucketName,
	).Scan(
		&b.ID, &b.ID, &b.Name, &b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *BucketStore) ListBuckets(ctx context.Context) ([]*Bucket, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, created_at
		FROM buckets
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []*Bucket
	for rows.Next() {
		var b Bucket
		if err := rows.Scan(&b.ID, &b.Name, &b.CreatedAt); err != nil {
			return nil, err
		}
		buckets = append(buckets, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return buckets, nil
}
