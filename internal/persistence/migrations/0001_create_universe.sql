-- Create the primary torrents table.
CREATE TABLE torrents (
    -- Core info.
    id INTEGER PRIMARY KEY,
    info_hash BLOB NOT NULL UNIQUE,
    name TEXT NOT NULL,
    total_size INTEGER NOT NULL,

    -- Timestamps.
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Optimize lookups for torrents by info hash.
CREATE UNIQUE INDEX info_hash_index ON torrents (info_hash);

-- Create the files table.
CREATE TABLE files (
    -- Core info.
    id INTEGER PRIMARY KEY,
    torrent_id INTEGER REFERENCES torrents ON DELETE CASCADE ON UPDATE RESTRICT,
    size INTEGER NOT NULL,
    path TEXT NOT NULL,
);

-- Optimize lookups for files by torrent ID.
CREATE INDEX files_torrent_id_index ON files (torrent_id);
