-- Create the primary torrents table.
CREATE TABLE torrents (
    id INTEGER PRIMARY KEY,
    info_hash BLOB NOT NULL UNIQUE,
    name TEXT NOT NULL,
    total_size INTEGER NOT NULL CHECK(total_size > 0),
    discovered_on INTEGER NOT NULL CHECK(discovered_on > 0)
);

-- Create the files table.
CREATE TABLE files (
    id INTEGER PRIMARY KEY,
    torrent_id INTEGER REFERENCES torrents ON DELETE CASCADE ON UPDATE RESTRICT,
    size INTEGER NOT NULL,
    path TEXT NOT NULL
);
