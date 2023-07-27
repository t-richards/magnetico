-- Create the primary torrents table.
CREATE TABLE torrents (
    -- Core info.
    id INTEGER PRIMARY KEY,
    info_hash BLOB NOT NULL UNIQUE,
    name TEXT NOT NULL,
    total_size INTEGER NOT NULL,

    -- Stats.
    seeder_count INTEGER NOT NULL DEFAULT 0,
    leecher_count INTEGER NOT NULL DEFAULT 0,

    -- Timestamps.
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE UNIQUE INDEX info_hash_index ON torrents (info_hash);

-- Create the files table.
CREATE TABLE files (
    -- Core info.
    id INTEGER PRIMARY KEY,
    torrent_id INTEGER REFERENCES torrents ON DELETE CASCADE ON UPDATE RESTRICT,
    size INTEGER NOT NULL,
    path TEXT NOT NULL,

    -- Readme file.
    is_readme INTEGER DEFAULT NULL,
    readme_content TEXT DEFAULT NULL
);

CREATE UNIQUE INDEX readme_index ON files (torrent_id, is_readme);
