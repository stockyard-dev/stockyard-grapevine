package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

// Article is a single help center / knowledge base article. Status is
// one of: published, draft, archived. The two counters track reader
// feedback (was this article helpful?).
type Article struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Category   string `json:"category"`
	Tags       string `json:"tags"`
	Slug       string `json:"slug"`
	Helpful    int    `json:"helpful"`
	NotHelpful int    `json:"not_helpful"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "grapevine.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS articles(
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		body TEXT DEFAULT '',
		category TEXT DEFAULT '',
		tags TEXT DEFAULT '',
		slug TEXT DEFAULT '',
		helpful INTEGER DEFAULT 0,
		not_helpful INTEGER DEFAULT 0,
		status TEXT DEFAULT 'published',
		created_at TEXT DEFAULT(datetime('now'))
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_category ON articles(category)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(
		resource TEXT NOT NULL,
		record_id TEXT NOT NULL,
		data TEXT NOT NULL DEFAULT '{}',
		PRIMARY KEY(resource, record_id)
	)`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string   { return time.Now().UTC().Format(time.RFC3339) }

func slugify(s string) string {
	out := strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range out {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	return b.String()
}

func (d *DB) Create(e *Article) error {
	e.ID = genID()
	e.CreatedAt = now()
	if e.Status == "" {
		e.Status = "published"
	}
	if e.Slug == "" {
		e.Slug = slugify(e.Title)
	}
	_, err := d.db.Exec(
		`INSERT INTO articles(id, title, body, category, tags, slug, helpful, not_helpful, status, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.Title, e.Body, e.Category, e.Tags, e.Slug, e.Helpful, e.NotHelpful, e.Status, e.CreatedAt,
	)
	return err
}

func (d *DB) Get(id string) *Article {
	var e Article
	err := d.db.QueryRow(
		`SELECT id, title, body, category, tags, slug, helpful, not_helpful, status, created_at
		 FROM articles WHERE id=?`,
		id,
	).Scan(&e.ID, &e.Title, &e.Body, &e.Category, &e.Tags, &e.Slug, &e.Helpful, &e.NotHelpful, &e.Status, &e.CreatedAt)
	if err != nil {
		return nil
	}
	return &e
}

// GetBySlug looks up an article by its URL slug. Useful for public
// help center routes.
func (d *DB) GetBySlug(slug string) *Article {
	var e Article
	err := d.db.QueryRow(
		`SELECT id, title, body, category, tags, slug, helpful, not_helpful, status, created_at
		 FROM articles WHERE slug=?`,
		slug,
	).Scan(&e.ID, &e.Title, &e.Body, &e.Category, &e.Tags, &e.Slug, &e.Helpful, &e.NotHelpful, &e.Status, &e.CreatedAt)
	if err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Article {
	rows, _ := d.db.Query(
		`SELECT id, title, body, category, tags, slug, helpful, not_helpful, status, created_at
		 FROM articles ORDER BY created_at DESC`,
	)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Article
	for rows.Next() {
		var e Article
		rows.Scan(&e.ID, &e.Title, &e.Body, &e.Category, &e.Tags, &e.Slug, &e.Helpful, &e.NotHelpful, &e.Status, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}

func (d *DB) Update(e *Article) error {
	_, err := d.db.Exec(
		`UPDATE articles SET title=?, body=?, category=?, tags=?, slug=?, helpful=?, not_helpful=?, status=?
		 WHERE id=?`,
		e.Title, e.Body, e.Category, e.Tags, e.Slug, e.Helpful, e.NotHelpful, e.Status, e.ID,
	)
	return err
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM articles WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&n)
	return n
}

// MarkHelpful atomically increments the helpful counter. Avoids the
// read-modify-write race that the original implementation would have
// had if the dashboard tried to use PUT.
func (d *DB) MarkHelpful(id string) (int, error) {
	_, err := d.db.Exec(`UPDATE articles SET helpful = helpful + 1 WHERE id=?`, id)
	if err != nil {
		return 0, err
	}
	var v int
	d.db.QueryRow(`SELECT helpful FROM articles WHERE id=?`, id).Scan(&v)
	return v, nil
}

func (d *DB) MarkNotHelpful(id string) (int, error) {
	_, err := d.db.Exec(`UPDATE articles SET not_helpful = not_helpful + 1 WHERE id=?`, id)
	if err != nil {
		return 0, err
	}
	var v int
	d.db.QueryRow(`SELECT not_helpful FROM articles WHERE id=?`, id).Scan(&v)
	return v, nil
}

func (d *DB) Search(q string, filters map[string]string) []Article {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (title LIKE ? OR body LIKE ? OR tags LIKE ?)"
		s := "%" + q + "%"
		args = append(args, s, s, s)
	}
	if v, ok := filters["category"]; ok && v != "" {
		where += " AND category=?"
		args = append(args, v)
	}
	if v, ok := filters["status"]; ok && v != "" {
		where += " AND status=?"
		args = append(args, v)
	}
	rows, _ := d.db.Query(
		`SELECT id, title, body, category, tags, slug, helpful, not_helpful, status, created_at
		 FROM articles WHERE `+where+`
		 ORDER BY helpful DESC, created_at DESC`,
		args...,
	)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []Article
	for rows.Next() {
		var e Article
		rows.Scan(&e.ID, &e.Title, &e.Body, &e.Category, &e.Tags, &e.Slug, &e.Helpful, &e.NotHelpful, &e.Status, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}

// Stats returns total articles, sums of both vote counters, by_status
// and by_category breakdowns.
func (d *DB) Stats() map[string]any {
	m := map[string]any{
		"total":           d.Count(),
		"total_helpful":   0,
		"total_unhelpful": 0,
		"by_status":       map[string]int{},
		"by_category":     map[string]int{},
	}

	var helpful, unhelpful int
	d.db.QueryRow(`SELECT COALESCE(SUM(helpful), 0) FROM articles`).Scan(&helpful)
	d.db.QueryRow(`SELECT COALESCE(SUM(not_helpful), 0) FROM articles`).Scan(&unhelpful)
	m["total_helpful"] = helpful
	m["total_unhelpful"] = unhelpful

	if rows, _ := d.db.Query(`SELECT status, COUNT(*) FROM articles GROUP BY status`); rows != nil {
		defer rows.Close()
		by := map[string]int{}
		for rows.Next() {
			var s string
			var c int
			rows.Scan(&s, &c)
			by[s] = c
		}
		m["by_status"] = by
	}

	if rows, _ := d.db.Query(`SELECT category, COUNT(*) FROM articles WHERE category != '' GROUP BY category`); rows != nil {
		defer rows.Close()
		by := map[string]int{}
		for rows.Next() {
			var s string
			var c int
			rows.Scan(&s, &c)
			by[s] = c
		}
		m["by_category"] = by
	}

	return m
}

// ─── Extras ───────────────────────────────────────────────────────

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
