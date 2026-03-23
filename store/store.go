package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Proposal struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	Title      string     `json:"title"`
	Summary    string     `json:"summary"`
	Plan       string     `json:"plan"`
	SessionID  string     `json:"session_id"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS proposals (
		id          TEXT PRIMARY KEY,
		status      TEXT NOT NULL DEFAULT 'pending',
		title       TEXT NOT NULL,
		summary     TEXT NOT NULL,
		plan        TEXT NOT NULL,
		session_id  TEXT NOT NULL,
		created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
		resolved_at DATETIME
	)`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) PendingCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM proposals WHERE status = 'pending'").Scan(&count)
	return count, err
}

func (s *Store) Insert(p *Proposal) error {
	_, err := s.db.Exec(
		"INSERT INTO proposals (id, title, summary, plan, session_id) VALUES (?, ?, ?, ?, ?)",
		p.ID, p.Title, p.Summary, p.Plan, p.SessionID,
	)
	return err
}

func (s *Store) Get(id string) (*Proposal, error) {
	p := &Proposal{}
	err := s.db.QueryRow(
		"SELECT id, status, title, summary, plan, session_id, created_at, resolved_at FROM proposals WHERE id = ?", id,
	).Scan(&p.ID, &p.Status, &p.Title, &p.Summary, &p.Plan, &p.SessionID, &p.CreatedAt, &p.ResolvedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) List(status string) ([]Proposal, error) {
	query := "SELECT id, status, title, summary, plan, session_id, created_at, resolved_at FROM proposals"
	var args []any
	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var proposals []Proposal
	for rows.Next() {
		var p Proposal
		if err := rows.Scan(&p.ID, &p.Status, &p.Title, &p.Summary, &p.Plan, &p.SessionID, &p.CreatedAt, &p.ResolvedAt); err != nil {
			return nil, err
		}
		proposals = append(proposals, p)
	}
	return proposals, rows.Err()
}

func (s *Store) Approve(id string) error {
	res, err := s.db.Exec(
		"UPDATE proposals SET status = 'approved', resolved_at = datetime('now') WHERE id = ? AND status = 'pending'",
		id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) Reject(id string) error {
	res, err := s.db.Exec(
		"UPDATE proposals SET status = 'rejected', resolved_at = datetime('now') WHERE id = ? AND status = 'pending'",
		id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
