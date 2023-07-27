CREATE VIRTUAL TABLE torrents_idx USING fts5(name, content='torrents', content_rowid='id');

CREATE TRIGGER torrents_idx_ai_t AFTER INSERT ON torrents BEGIN
    INSERT INTO torrents_idx(rowid, name) VALUES (new.id, new.name);
END;

CREATE TRIGGER torrents_idx_ad_t AFTER DELETE ON torrents BEGIN
    INSERT INTO torrents_idx(torrents_idx, rowid, name) VALUES('delete', old.id, old.name);
END;

CREATE TRIGGER torrents_idx_au_t AFTER UPDATE ON torrents BEGIN
    INSERT INTO torrents_idx(torrents_idx, rowid, name) VALUES('delete', old.id, old.name);
    INSERT INTO torrents_idx(rowid, name) VALUES (new.id, new.name);
END;