package toki

import "database/sql"

// Stmt represents a prepared SQL statement
type Stmt struct {
	query string
	args  []interface{}
	db    *sql.DB
	tx    *sql.Tx
}

// Prepare creates a prepared statement
func (b *Builder) Prepare(db *sql.DB) (*Stmt, error) {
	query := b.String()

	stmt := &Stmt{
		query: query,
		args:  b.args,
		db:    db,
	}

	if b.tx != nil {
		stmt.tx = b.tx.tx
	}

	return stmt, nil
}

// Query executes the query and returns rows
func (s *Stmt) Query() (*sql.Rows, error) {
	if s.tx != nil {
		return s.tx.Query(s.query, s.args...)
	}
	return s.db.Query(s.query, s.args...)
}

// QueryRow executes the query and returns a single row
func (s *Stmt) QueryRow() *sql.Row {
	if s.tx != nil {
		return s.tx.QueryRow(s.query, s.args...)
	}
	return s.db.QueryRow(s.query, s.args...)
}

// Exec executes the statement
func (s *Stmt) Exec() (sql.Result, error) {
	if s.tx != nil {
		return s.tx.Exec(s.query, s.args...)
	}
	return s.db.Exec(s.query, s.args...)
}
