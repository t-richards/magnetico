CREATE VIRTUAL TABLE torrents_idx USING fts5(name, content='torrents', content_rowid='id', tokenize="porter unicode61 separators ' !""#$%&''()*+,-./:;<=>?@[\]^_`{|}~'");

-- Populate the index
INSERT INTO torrents_idx(rowid, name) SELECT id, name FROM torrents;

-- Triggers to keep the FTS index up to date.
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

-- Add column 'modified_on'
-- BEWARE: code needs to be updated before January 1, 3000 (32503680000)!
ALTER TABLE torrents ADD COLUMN modified_on INTEGER NOT NULL
    CHECK (modified_on >= discovered_on AND (updated_on IS NOT NULL OR modified_on >= updated_on))
    DEFAULT 32503680000
;

-- If 'modified_on' is not explicitly supplied, then it shall be set, by default, to
-- 'discovered_on' right after the row is inserted to 'torrents'.
--
-- {WHEN expr} does NOT work for some reason (trigger doesn't get triggered), so we use
--   AND NEW."modified_on" = 32503680000
-- instead in the WHERE clause.
CREATE TRIGGER "torrents_modified_on_default_t" AFTER INSERT ON "torrents" BEGIN
    UPDATE "torrents" SET "modified_on" = NEW."discovered_on" WHERE "id" = NEW."id" AND NEW."modified_on" = 32503680000;
END;

-- Set 'modified_on' value of all rows to 'discovered_on' or 'updated_on', whichever is
-- greater; beware that 'updated_on' can be NULL too.
UPDATE torrents SET modified_on = (SELECT MAX(discovered_on, IFNULL(updated_on, 0)));

CREATE INDEX modified_on_index ON torrents (modified_on);
