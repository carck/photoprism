CREATE VIRTUAL TABLE photo_search USING fts5(keywords, notes, content='', tokenize = 'simple');

-- Triggers to keep the FTS index up to date.
CREATE TRIGGER photos_ai AFTER INSERT ON details BEGIN
  INSERT INTO photo_search(rowid, keywords, notes) VALUES (new.photo_id, new.keywords, new.notes);
END;
CREATE TRIGGER photos_ad AFTER DELETE ON details BEGIN
  INSERT INTO photo_search(photo_search, rowid, keywords, notes) VALUES('delete', old.photo_id, old.keywords, old.notes);
END;
CREATE TRIGGER photos_au AFTER UPDATE ON details BEGIN
  INSERT INTO photo_search(photo_search, rowid, keywords, notes) VALUES('delete', old.photo_id, old.keywords, old.notes);
  INSERT INTO photo_search(rowid, keywords, notes) VALUES (new.photo_id, new.keywords, new.notes);
END;

INSERT INTO photo_search(rowid, keywords, notes) select photo_id, keywords, notes from details;