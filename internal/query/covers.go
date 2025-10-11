package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
)

// UpdateAlbumDefaultCovers updates default album cover thumbs.
func UpdateAlbumDefaultCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("album_type = ? AND thumb_src = ?", entity.AlbumDefault, entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE albums LEFT JOIN (
    	SELECT p2.album_uid, f.file_hash FROM files f, (
        	SELECT pa.album_uid, max(p.id) AS photo_id FROM photos p
            JOIN photos_albums pa ON pa.photo_uid = p.photo_uid AND pa.hidden = 0 AND pa.missing = 0
        	WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
        	GROUP BY pa.album_uid) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type in ('jpg','heif')
			) b ON b.album_uid = albums.album_uid
		SET thumb = b.file_hash WHERE ?`, condition)
	case SQLite3:
		res = Db().Exec(`
		WITH latest_album AS (
			SELECT pa.album_uid, MAX(f.id) AS file_id
			FROM photos_albums pa
			JOIN files f ON f.photo_uid = pa.photo_uid
			JOIN photos p ON p.id = f.photo_id
			WHERE pa.hidden = 0 AND pa.missing = 0
				AND f.deleted_at IS NULL
				AND f.file_missing = 0
				AND f.file_primary = 1
				AND f.file_hash <> ''
				AND f.file_error = ''
				AND f.file_type IN ('jpg','heif')
				AND p.photo_private = 0
				AND p.deleted_at IS NULL
				AND p.photo_quality > 0
			GROUP BY pa.album_uid
			),
			chosen_files AS (
			SELECT la.album_uid, f.file_hash
			FROM latest_album la
			JOIN files f ON f.id = la.file_id
			)
			UPDATE albums
			SET thumb = (
			SELECT cf.file_hash
			FROM chosen_files cf
			WHERE cf.album_uid = albums.album_uid
			)
			WHERE album_type = ? AND thumb_src = ?;
		`, entity.AlbumDefault, entity.SrcAuto)
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "album", "albums"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating albums, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateAlbumFolderCovers updates folder album cover thumbs.
func UpdateAlbumFolderCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("album_type = ? AND thumb_src = ?", entity.AlbumFolder, entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE albums LEFT JOIN (
		SELECT p2.photo_path, f.file_hash FROM files f, (
			SELECT p.photo_path, max(p.id) AS photo_id FROM photos p
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY p.photo_path) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type  in ('jpg','heif')
			) b ON b.photo_path = albums.album_path
		SET thumb = b.file_hash WHERE ?`, condition)
	case SQLite3:
		res = Db().Exec(`
		WITH chosen_files AS (
			SELECT p.photo_path, MAX(f.id) AS file_id
			FROM files f
			JOIN photos p ON f.photo_id = p.id
			WHERE f.file_primary = 1
			AND f.file_error = ''
			AND f.file_type IN ('jpg','heif')
			GROUP BY p.photo_path
		),
		chosen_hashes AS (
			SELECT cf.photo_path, f.file_hash
			FROM chosen_files cf
			JOIN files f ON f.id = cf.file_id
		)
		UPDATE albums
		SET thumb = (
			SELECT ch.file_hash 
			FROM chosen_hashes ch
			WHERE ch.photo_path = albums.album_path
		)
		WHERE ?;
		`, condition)

	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "folder", "folders"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating folders, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateAlbumMonthCovers updates month album cover thumbs.
func UpdateAlbumMonthCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("album_type = ? AND thumb_src = ?", entity.AlbumMonth, entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE albums LEFT JOIN (
		SELECT p2.photo_year, p2.photo_month, f.file_hash FROM files f, (
			SELECT p.photo_year, p.photo_month, max(p.id) AS photo_id FROM photos p
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY p.photo_year, p.photo_month) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type  in ('jpg','heif')
			) b ON b.photo_year = albums.album_year AND b.photo_month = albums.album_month
		SET thumb = b.file_hash WHERE ?`, condition)
	case SQLite3:
		res = Db().Exec(`
			WITH chosen_files AS (
				SELECT b.photo_year, b.photo_month, MAX(f.id) AS file_id
				FROM photos b
				JOIN files f ON f.photo_id = b.id
				WHERE f.file_primary = 1
				AND f.file_error = ''
				AND f.file_type IN ('jpg','heif')
				GROUP BY b.photo_year, b.photo_month
			),
			chosen_hashes AS (
				SELECT cf.photo_year, cf.photo_month, f.file_hash
				FROM chosen_files cf
				JOIN files f ON f.id = cf.file_id
			)
			UPDATE albums
			SET thumb = (
				SELECT ch.file_hash 
				FROM chosen_hashes ch
				WHERE ch.photo_year = albums.album_year
				AND ch.photo_month = albums.album_month
			)
			WHERE ?;
			`, condition)

	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "month", "months"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating calendar, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateAlbumCovers updates album cover thumbs.
func UpdateAlbumCovers() (err error) {
	// Update Default Albums.
	if err = UpdateAlbumDefaultCovers(); err != nil {
		return err
	}

	// Update Folder Albums.
	if err = UpdateAlbumFolderCovers(); err != nil {
		return err
	}

	// Update Monthly Albums.
	if err = UpdateAlbumMonthCovers(); err != nil {
		return err
	}

	return nil
}

// UpdateLabelCovers updates label cover thumbs.
func UpdateLabelCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	condition := gorm.Expr("thumb_src = ?", entity.SrcAuto)

	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE labels LEFT JOIN (
		SELECT p2.label_id, f.file_hash FROM files f, (
			SELECT pl.label_id as label_id, max(p.id) AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY pl.label_id
			UNION
			SELECT c.category_id as label_id, max(p.id) AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
				JOIN categories c ON c.label_id = pl.label_id
			WHERE p.photo_quality > 0 AND p.photo_private = 0 AND p.deleted_at IS NULL
			GROUP BY c.category_id
			) p2 WHERE p2.photo_id = f.photo_id AND f.file_primary = 1 AND f.file_error = '' AND f.file_type  in ('jpg','heif') AND f.file_missing = 0
		) b ON b.label_id = labels.id
		SET thumb = b.file_hash WHERE ?`, condition)
	case SQLite3:
		res = Db().Exec(`
		WITH label_files AS (
			SELECT pl.label_id, MAX(f.id) AS file_id
			FROM photos_labels pl
			JOIN files f ON f.photo_id = pl.photo_id
			WHERE pl.uncertainty < 21
				AND f.deleted_at IS NULL
				AND f.file_hash <> ''
				AND f.file_missing = 0
				AND f.file_primary = 1
				AND f.file_type IN ('jpg','heif')
			GROUP BY pl.label_id
			),
			label_hashes AS (
			SELECT lf.label_id, f.file_hash
			FROM label_files lf
			JOIN files f ON f.id = lf.file_id
			)
			UPDATE labels
			SET thumb = (
			SELECT lh.file_hash
			FROM label_hashes lh
			WHERE lh.label_id = labels.id
			)
			WHERE thumb_src = ?;
		`, entity.SrcAuto)

		/**
		if res.Error == nil {
			catRes := Db().Exec(`
				WITH category_files AS (
					SELECT c.category_id AS label_id,
							MAX(pl.photo_id) AS photo_id
					FROM categories c
					JOIN photos_labels pl ON pl.label_id = c.label_id
					WHERE pl.uncertainty < 21
					GROUP BY c.category_id
					),
					category_hashes AS (
					SELECT cf.label_id, f.file_hash
					FROM category_files cf
					JOIN files f ON f.photo_id = cf.photo_id
					WHERE f.file_hash <> ''
						AND f.deleted_at IS NULL
						AND f.file_hash <> ''
						AND f.file_missing = 0
						AND f.file_primary = 1
						AND f.file_type IN ('jpg','heif')
					)
					UPDATE labels
					SET thumb = (
						SELECT ch.file_hash
						FROM category_hashes ch
						WHERE ch.label_id = labels.id
					)
					WHERE thumb IS NULL OR thumb = '';
			`)

			res.RowsAffected += catRes.RowsAffected
		}*/
	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "label", "labels"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating labels, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateSubjectCovers updates subject cover thumbs.
func UpdateSubjectCovers() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	subjTable := entity.Subject{}.TableName()
	markerTable := entity.Marker{}.TableName()

	condition := gorm.Expr(
		fmt.Sprintf("%s.subj_type = ? AND thumb_src = ?", subjTable),
		entity.SubjPerson, entity.SrcAuto)

	// TODO: Avoid using private photos as subject covers.
	switch DbDialect() {
	case MySQL:
		res = Db().Exec(`UPDATE ? LEFT JOIN (
    	SELECT m.subj_uid, m.q, MAX(m.thumb) AS marker_thumb FROM ? m
			WHERE m.subj_uid <> '' AND m.subj_uid IS NOT NULL
			  AND m.marker_invalid = 0 AND m.thumb IS NOT NULL AND m.thumb <> ''
			GROUP BY m.subj_uid, m.q
			) b ON b.subj_uid = subjects.subj_uid
		SET thumb = marker_thumb WHERE ?`, gorm.Expr(subjTable), gorm.Expr(markerTable), condition)
	case SQLite3:
		res = Db().Exec(`
			WITH ranked_markers AS (
				SELECT m.subj_uid, MAX(m.rowid) AS marker_id
				FROM markers m
				WHERE m.subj_uid <> '' and m.thumb <> '' and m.size > 80
				GROUP BY m.subj_uid
			),
			chosen_thumbs AS (
				SELECT rm.subj_uid, m.thumb
				FROM ranked_markers rm
				JOIN markers m ON m.rowid = rm.marker_id
			)
			UPDATE subjects
			SET thumb = (
				SELECT ct.thumb
				FROM chosen_thumbs ct
				WHERE ct.subj_uid = subjects.subj_uid
			)
			WHERE ?;
			`, condition)

	default:
		log.Warnf("sql: unsupported dialect %s", DbDialect())
		return nil
	}

	err = res.Error

	if err == nil {
		log.Debugf("covers: updated %s [%s]", english.Plural(int(res.RowsAffected), "subject", "subjects"), time.Since(start))
	} else if strings.Contains(err.Error(), "Error 1054") {
		log.Errorf("covers: failed updating subjects, potentially incompatible database version")
		log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
		return nil
	}

	return err
}

// UpdateCovers updates album, subject, and label cover thumbs.
func UpdateCovers() (err error) {
	log.Debugf("index: updating covers")

	// Update Albums.
	if err = UpdateAlbumCovers(); err != nil {
		return err
	}

	// Update Labels.
	if err = UpdateLabelCovers(); err != nil {
		return err
	}

	// Update Subjects.
	if err = UpdateSubjectCovers(); err != nil {
		return err
	}

	return nil
}
