CREATE VIRTUAL TABLE photo_search USING fts5(keywords, notes, content='', tokenize = 'simple', contentless_delete=1);

-- Triggers to keep the FTS index up to date.
CREATE TRIGGER photos_ai AFTER INSERT ON details BEGIN
  INSERT INTO photo_search(rowid, keywords, notes) VALUES (new.photo_id, new.keywords, new.notes);
END;
CREATE TRIGGER photos_ad AFTER DELETE ON details BEGIN
  delete from photo_search where rowid = old.photo_id;
END;
CREATE TRIGGER photos_au AFTER UPDATE ON details BEGIN
  update photo_search set keywords = new.keywords, notes = new.notes where rowid = new.photo_id;
END;

INSERT INTO photo_search(rowid, keywords, notes) select photo_id, keywords, notes from details;
