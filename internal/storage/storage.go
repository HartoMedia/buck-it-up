package storage

// storage.go
// Helpers for object file system layout.

import (
	"os"
	"path/filepath"
	"strconv"
)

// DataRoot returns root directory for storing bucket/object files.
// Controlled by BUCK_DATA_PATH env var, defaults to "data".
func DataRoot() string {
	if dr := os.Getenv("BUCK_DATA_PATH"); dr != "" {
		return dr
	}
	return "data"
}

// EnsureDataRoot makes sure the data root exists.
func EnsureDataRoot() error {
	return os.MkdirAll(DataRoot(), 0o755)
}

// EnsureBucketObjectsDir creates the directory that will hold object files for a bucket.
func EnsureBucketObjectsDir(bucketID int64) (string, error) {
	p := filepath.Join(DataRoot(), "buckets", strconv.FormatInt(bucketID, 10), "objects")
	return p, os.MkdirAll(p, 0o755)
}

// ObjectFilePath returns the file path for a given object ID in a bucket.
// This should match the format stored in the DB ("data/buckets/<bucket_id>/objects/<id>").
func ObjectFilePath(bucketID, objectID int64) string {
	return filepath.Join(DataRoot(), "buckets", strconv.FormatInt(bucketID, 10), "objects", strconv.FormatInt(objectID, 10))
}
