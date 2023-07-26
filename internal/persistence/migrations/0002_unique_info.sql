-- Make info hashes unique.
CREATE UNIQUE INDEX info_hash_index ON torrents	(info_hash);
