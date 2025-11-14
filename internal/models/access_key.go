// internal/models/access_keys.go

package models

// access_keys.go

import (
	"context"
	"database/sql"
)

type AccessKeyStore struct {
	db *sql.DB
}

func NewAccessKeyStore(db *sql.DB) *AccessKeyStore {
	return &AccessKeyStore{db: db}
}

func (s *AccessKeyStore) GetByKeyID(ctx context.Context, keyID string) (*AccessKey, error) {
	var ak AccessKey
	err := s.db.QueryRowContext(ctx, `
        SELECT id, bucket_id, key_id, secret_hash, role, created_at
        FROM access_keys
        WHERE key_id = ?
    `, keyID).Scan(
		&ak.ID,
		&ak.BucketID,
		&ak.KeyID,
		&ak.SecretHash,
		&ak.Role,
		&ak.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ak, nil
}

func (s *AccessKeyStore) CreateAccessKey(ctx context.Context, ak *AccessKey) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO access_keys (
			bucket_id, key_id, secret_hash, role, created_at
		) VALUES (?, ?, ?, ?, ?)
	`,
		ak.BucketID,
		ak.KeyID,
		ak.SecretHash,
		ak.Role,
		ak.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *AccessKeyStore) DeleteAccessKeyByBucketAndRole(ctx context.Context, bucketID int64, role AccessKeyRole) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM access_keys
		WHERE bucket_id = ? AND role = ?
	`, bucketID, role)
	return err
}

func (s *AccessKeyStore) ListAccessKeysByBucketID(ctx context.Context, bucketID int64) ([]*AccessKey, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, bucket_id, key_id, secret_hash, role, created_at
		FROM access_keys
		WHERE bucket_id = ?
		ORDER BY role
	`, bucketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*AccessKey
	for rows.Next() {
		var ak AccessKey
		err := rows.Scan(
			&ak.ID,
			&ak.BucketID,
			&ak.KeyID,
			&ak.SecretHash,
			&ak.Role,
			&ak.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, &ak)
	}
	return keys, rows.Err()
}
