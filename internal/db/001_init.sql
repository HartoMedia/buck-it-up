CREATE TABLE buckets (
                         id          INTEGER PRIMARY KEY AUTOINCREMENT,
                         name        TEXT NOT NULL UNIQUE,
                         created_at  INTEGER NOT NULL
);

CREATE TABLE access_keys (
                             id           INTEGER PRIMARY KEY AUTOINCREMENT,
                             bucket_id    INTEGER NOT NULL,
                             key_id       TEXT NOT NULL,          -- public part
                             secret_hash  TEXT NOT NULL,          -- hashed secret
                             role         TEXT NOT NULL,          -- 'readOnly', 'readWrite', 'all'
                             created_at   INTEGER NOT NULL,
                             FOREIGN KEY(bucket_id) REFERENCES buckets(id),
                             UNIQUE (key_id)
);

CREATE TABLE objects (
                         id            INTEGER PRIMARY KEY AUTOINCREMENT,
                         bucket_id     INTEGER NOT NULL,
                         object_key    TEXT NOT NULL,       -- "folder/file.txt"
                         file_path     TEXT NOT NULL,       -- "data/buckets/<bucket_id>/objects/<id>"
                         size          INTEGER NOT NULL,
                         content_type  TEXT,
                         checksum      TEXT,                -- optional, e.g. SHA256
                         created_at    INTEGER NOT NULL,
                         FOREIGN KEY(bucket_id) REFERENCES buckets(id),
                         UNIQUE(bucket_id, object_key)
);


CREATE INDEX idx_access_keys_bucket_id ON access_keys(bucket_id);
CREATE INDEX idx_objects_bucket_id ON objects(bucket_id);
CREATE INDEX idx_objects_bucket_id_object_key ON objects(bucket_id, object_key);

