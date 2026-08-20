package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"api/internal/models"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	// SQLite has a single writer; one connection serializes access and
	// avoids SQLITE_BUSY errors under concurrent requests.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS media (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    type         TEXT NOT NULL CHECK(type IN ('photo','video')),
    source       TEXT NOT NULL CHECK(source IN ('upload','youtube','vimeo')),
    path         TEXT NOT NULL DEFAULT '',
    external_id  TEXT NOT NULL DEFAULT '',
    title        TEXT NOT NULL DEFAULT '',
    instagram_url TEXT NOT NULL DEFAULT '',
    description  TEXT NOT NULL DEFAULT '',
    youtube_url  TEXT NOT NULL DEFAULT '',
    vimeo_url    TEXT NOT NULL DEFAULT '',
    preview_path TEXT NOT NULL DEFAULT '',
    width        INTEGER NOT NULL DEFAULT 0,
    height       INTEGER NOT NULL DEFAULT 0,
    featured     INTEGER NOT NULL DEFAULT 0,
    position     INTEGER NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`); err != nil {
		return err
	}
	// add columns to databases created before they existed
	for _, col := range []struct{ name, ddl string }{
		{"featured", `ALTER TABLE media ADD COLUMN featured INTEGER NOT NULL DEFAULT 0`},
		{"preview_path", `ALTER TABLE media ADD COLUMN preview_path TEXT NOT NULL DEFAULT ''`},
		{"width", `ALTER TABLE media ADD COLUMN width INTEGER NOT NULL DEFAULT 0`},
		{"height", `ALTER TABLE media ADD COLUMN height INTEGER NOT NULL DEFAULT 0`},
		{"description", `ALTER TABLE media ADD COLUMN description TEXT NOT NULL DEFAULT ''`},
		{"youtube_url", `ALTER TABLE media ADD COLUMN youtube_url TEXT NOT NULL DEFAULT ''`},
		{"vimeo_url", `ALTER TABLE media ADD COLUMN vimeo_url TEXT NOT NULL DEFAULT ''`},
	} {
		has, err := hasColumn(db, "media", col.name)
		if err != nil {
			return err
		}
		if !has {
			if _, err := db.Exec(col.ddl); err != nil {
				return err
			}
		}
	}
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS translations (
    entity    TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    field     TEXT NOT NULL,
    lang      TEXT NOT NULL,
    value     TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (entity, entity_id, field, lang)
);`)
	return err
}

func hasColumn(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Ping reports whether the database is reachable; the health endpoint relies
// on it.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

const selectCols = `id, type, source, path, external_id, title, description, instagram_url, youtube_url, vimeo_url, preview_path, width, height, featured, position, created_at`

func scanMedia(row interface{ Scan(...any) error }) (*models.Media, error) {
	var m models.Media
	err := row.Scan(&m.ID, &m.Type, &m.Source, &m.Path, &m.ExternalID, &m.Title, &m.Description, &m.InstagramURL, &m.YoutubeURL, &m.VimeoURL, &m.PreviewPath, &m.Width, &m.Height, &m.Featured, &m.Position, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) ListMedia(ctx context.Context) ([]models.Media, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+selectCols+` FROM media ORDER BY position ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]models.Media, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *m)
		ids = append(ids, m.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	tr, err := s.LoadTranslations(ctx, "media", ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if t, ok := tr[items[i].ID]; ok {
			items[i].Translations = t
		} else {
			items[i].Translations = map[string]map[string]string{}
		}
	}
	return items, nil
}

func (s *Store) GetMedia(ctx context.Context, id int64) (*models.Media, error) {
	m, err := scanMedia(s.db.QueryRowContext(ctx,
		`SELECT `+selectCols+` FROM media WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	tr, err := s.LoadTranslations(ctx, "media", []int64{id})
	if err != nil {
		return nil, err
	}
	if t, ok := tr[id]; ok {
		m.Translations = t
	} else {
		m.Translations = map[string]map[string]string{}
	}
	return m, nil
}

func (s *Store) CreateMedia(ctx context.Context, m *models.Media) (*models.Media, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx,
		`INSERT INTO media (type, source, path, external_id, title, description, instagram_url, youtube_url, vimeo_url, preview_path, width, height, featured, position)
		 SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(MAX(position), 0) + 1 FROM media`,
		m.Type, m.Source, m.Path, m.ExternalID, m.Title, m.Description, m.InstagramURL, m.YoutubeURL, m.VimeoURL, m.PreviewPath, m.Width, m.Height, m.Featured)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetMedia(ctx, id)
}

// MediaPatch carries optional field updates; nil fields stay untouched.
type MediaPatch struct {
	Title        *string
	Description  *string
	InstagramURL *string
	YoutubeURL   *string
	VimeoURL     *string
	PreviewPath  *string
	Featured     *bool
}

func (s *Store) UpdateMedia(ctx context.Context, id int64, p MediaPatch) (*models.Media, error) {
	cols := make([]string, 0, 7)
	args := make([]any, 0, 8)
	addStr := func(col string, v *string) {
		if v != nil {
			cols = append(cols, col)
			args = append(args, *v)
		}
	}
	addStr("title", p.Title)
	addStr("description", p.Description)
	addStr("instagram_url", p.InstagramURL)
	addStr("youtube_url", p.YoutubeURL)
	addStr("vimeo_url", p.VimeoURL)
	addStr("preview_path", p.PreviewPath)
	if p.Featured != nil {
		cols = append(cols, "featured")
		args = append(args, *p.Featured)
	}
	if len(cols) > 0 {
		var q strings.Builder
		q.WriteString("UPDATE media SET ")
		for i, col := range cols {
			if i > 0 {
				q.WriteString(", ")
			}
			q.WriteString(col + " = ?")
		}
		q.WriteString(" WHERE id = ?")
		args = append(args, id)
		if _, err := s.db.ExecContext(ctx, q.String(), args...); err != nil {
			return nil, err
		}
	}
	return s.GetMedia(ctx, id)
}

// SetTranslations upserts per field/lang values for an entity; empty values
// delete the row so cleared translations actually fall back.
func (s *Store) SetTranslations(ctx context.Context, entity string, entityID int64, tr map[string]map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := setTranslations(ctx, tx, entity, entityID, tr); err != nil {
		return err
	}
	return tx.Commit()
}

func setTranslations(ctx context.Context, tx *sql.Tx, entity string, entityID int64, tr map[string]map[string]string) error {
	for field, langs := range tr {
		for lang, val := range langs {
			if val == "" {
				if _, err := tx.ExecContext(ctx,
					`DELETE FROM translations WHERE entity = ? AND entity_id = ? AND field = ? AND lang = ?`,
					entity, entityID, field, lang); err != nil {
					return err
				}
				continue
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO translations (entity, entity_id, field, lang, value) VALUES (?, ?, ?, ?, ?)
				 ON CONFLICT(entity, entity_id, field, lang) DO UPDATE SET value = excluded.value`,
				entity, entityID, field, lang, val); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadTranslations returns entityID → field → lang → value for the given ids.
func (s *Store) LoadTranslations(ctx context.Context, entity string, ids []int64) (map[int64]map[string]map[string]string, error) {
	out := make(map[int64]map[string]map[string]string)
	if len(ids) == 0 {
		return out, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, entity)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx,
		// #nosec G202 -- the concatenated part is only literal "?"
		// placeholders; all values stay parameterized.
		`SELECT entity_id, field, lang, value FROM translations
		 WHERE entity = ? AND entity_id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id int64
		var field, lang, val string
		if err := rows.Scan(&id, &field, &lang, &val); err != nil {
			return nil, err
		}
		if out[id] == nil {
			out[id] = make(map[string]map[string]string)
		}
		if out[id][field] == nil {
			out[id][field] = make(map[string]string)
		}
		out[id][field][lang] = val
	}
	return out, rows.Err()
}

// GetSettings returns all stored key/value settings.
func (s *Store) GetSettings(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, rows.Err()
}

// SetSettings upserts the given key/value pairs in one transaction.
func (s *Store) SetSettings(ctx context.Context, kv map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := setSettings(ctx, tx, kv); err != nil {
		return err
	}
	return tx.Commit()
}

func setSettings(ctx context.Context, tx *sql.Tx, kv map[string]string) error {
	for k, v := range kv {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO settings (key, value) VALUES (?, ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value`, k, v); err != nil {
			return err
		}
	}
	return nil
}

// SetSettingsAndTranslations upserts settings pairs and the settings
// translations (entity "settings", id 0) atomically, so a failed save leaves
// neither half applied.
func (s *Store) SetSettingsAndTranslations(ctx context.Context, kv map[string]string, tr map[string]map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := setSettings(ctx, tx, kv); err != nil {
		return err
	}
	if err := setTranslations(ctx, tx, "settings", 0, tr); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteMedia removes the row along with its translations and returns it (so
// the caller can clean up uploaded files). Returns nil, nil when the id does
// not exist.
func (s *Store) DeleteMedia(ctx context.Context, id int64) (*models.Media, error) {
	m, err := s.GetMedia(ctx, id)
	if err != nil || m == nil {
		return m, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM translations WHERE entity = 'media' AND entity_id = ?`, id); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM media WHERE id = ?`, id); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return m, nil
}

// Reorder sets position = index for each id. ids must be exactly the set of
// existing media ids, otherwise an error is returned and nothing changes.
// The read and the updates run in one transaction so a concurrent change
// cannot slip between validation and write.
func (s *Store) Reorder(ctx context.Context, ids []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Read ids inside a closure so rows are closed (via defer) before the
	// updates below run on the same transaction.
	existing, err := func() (map[int64]bool, error) {
		rows, err := tx.QueryContext(ctx, `SELECT id FROM media`)
		if err != nil {
			return nil, err
		}
		defer func() { _ = rows.Close() }()
		existing := make(map[int64]bool)
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			existing[id] = true
		}
		return existing, rows.Err()
	}()
	if err != nil {
		return err
	}
	if len(ids) != len(existing) {
		return fmt.Errorf("ids count %d does not match media count %d", len(ids), len(existing))
	}
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if !existing[id] {
			return fmt.Errorf("unknown media id %d", id)
		}
		if seen[id] {
			return fmt.Errorf("duplicate media id %d", id)
		}
		seen[id] = true
	}

	for i, id := range ids {
		if _, err := tx.ExecContext(ctx, `UPDATE media SET position = ? WHERE id = ?`, i, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
