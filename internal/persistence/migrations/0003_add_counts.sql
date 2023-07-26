-- Add columns for the number of seeders and leechers.
ALTER TABLE torrents ADD COLUMN updated_on INTEGER CHECK (updated_on > 0) DEFAULT NULL;
ALTER TABLE torrents ADD COLUMN n_seeders  INTEGER CHECK ((updated_on IS NOT NULL AND n_seeders >= 0) OR (updated_on IS NULL AND n_seeders IS NULL)) DEFAULT NULL;
ALTER TABLE torrents ADD COLUMN n_leechers INTEGER CHECK ((updated_on IS NOT NULL AND n_leechers >= 0) OR (updated_on IS NULL AND n_leechers IS NULL)) DEFAULT NULL;

-- Add columns for the contents of the README file.
ALTER TABLE files ADD COLUMN is_readme INTEGER CHECK (is_readme IS NULL OR is_readme=1) DEFAULT NULL;
ALTER TABLE files ADD COLUMN content   TEXT    CHECK ((content IS NULL AND is_readme IS NULL) OR (content IS NOT NULL AND is_readme=1)) DEFAULT NULL;
CREATE UNIQUE INDEX readme_index ON files (torrent_id, is_readme);
