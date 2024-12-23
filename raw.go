package toki

import "database/sql"

// RawQuery represents a raw SQL query
type RawQuery struct {
	sql  string
	args []interface{}
	db   *sql.DB
	tx   *sql.Tx
}

// Raw creates a new raw SQL query
func (b *Builder) Raw(sql string, args ...interface{}) *RawQuery {
	return &RawQuery{
		sql:  sql,
		args: args,
	}
}

// WithDB sets the database connection
func (r *RawQuery) WithDB(db *sql.DB) *RawQuery {
	r.db = db
	return r
}

// WithTx sets the transaction
func (r *RawQuery) WithTx(tx *sql.Tx) *RawQuery {
	r.tx = tx
	return r
}

// Query executes the raw query and returns rows
func (r *RawQuery) Query() (*sql.Rows, error) {
	if r.tx != nil {
		return r.tx.Query(r.sql, r.args...)
	}
	return r.db.Query(r.sql, r.args...)
}

// QueryRow executes the raw query and returns a single row
func (r *RawQuery) QueryRow() *sql.Row {
	if r.tx != nil {
		return r.tx.QueryRow(r.sql, r.args...)
	}
	return r.db.QueryRow(r.sql, r.args...)
}

// Exec executes the raw query
func (r *RawQuery) Exec() (sql.Result, error) {
	if r.tx != nil {
		return r.tx.Exec(r.sql, r.args...)
	}
	return r.db.Exec(r.sql, r.args...)
}

// String returns the SQL query string
func (r *RawQuery) String() string {
	return r.sql
}

// Args returns the query arguments
func (r *RawQuery) Args() []interface{} {
	return r.args
}
