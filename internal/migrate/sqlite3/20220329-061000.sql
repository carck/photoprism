DROP INDEX IF EXISTS idx_files_photo_id;
CREATE INDEX IF NOT EXISTS idx_files_photo_id ON files (photo_id, file_primary);