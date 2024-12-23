package toki

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRawQuery(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Test cases
	tests := []struct {
		name     string
		sql      string
		args     []interface{}
		setup    func(sqlmock.Sqlmock)
		validate func(*testing.T, *RawQuery)
		wantErr  bool
	}{
		{
			name: "Simple SELECT query",
			sql:  "SELECT * FROM users WHERE id = $1",
			args: []interface{}{1},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM users WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
						AddRow(1, "zakirkun"))
			},
			validate: func(t *testing.T, q *RawQuery) {
				rows, err := q.Query()
				assert.NoError(t, err)
				defer rows.Close()

				var id int
				var name string
				assert.True(t, rows.Next())
				err = rows.Scan(&id, &name)
				assert.NoError(t, err)
				assert.Equal(t, 1, id)
				assert.Equal(t, "zakirkun", name)
			},
		},
		{
			name: "INSERT with returning",
			sql:  "INSERT INTO users(name, email) VALUES($1, $2) RETURNING id",
			args: []interface{}{"zakirkun", "zakir@example.com"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users.*RETURNING id").
					WithArgs("zakirkun", "zakir@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			validate: func(t *testing.T, q *RawQuery) {
				var id int
				err := q.QueryRow().Scan(&id)
				assert.NoError(t, err)
				assert.Equal(t, 1, id)
			},
		},
		{
			name: "UPDATE query",
			sql:  "UPDATE users SET last_login = $1 WHERE id = $2",
			args: []interface{}{time.Now(), 1},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE users SET last_login = \\$1 WHERE id = \\$2").
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			validate: func(t *testing.T, q *RawQuery) {
				result, err := q.Exec()
				assert.NoError(t, err)
				affected, err := result.RowsAffected()
				assert.NoError(t, err)
				assert.Equal(t, int64(1), affected)
			},
		},
		{
			name: "Complex query with joins",
			sql: `
                SELECT u.*, p.name as profile_name 
                FROM users u 
                LEFT JOIN profiles p ON p.user_id = u.id 
                WHERE u.created_at > $1 AND u.status = $2
            `,
			args: []interface{}{time.Now().Add(-24 * time.Hour), "active"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT u.\\*, p\\.name as profile_name.*").
					WithArgs(sqlmock.AnyArg(), "active").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "profile_name"}).
						AddRow(1, "zakirkun", "Developer"))
			},
			validate: func(t *testing.T, q *RawQuery) {
				rows, err := q.Query()
				assert.NoError(t, err)
				defer rows.Close()

				var id int
				var name, profileName string
				assert.True(t, rows.Next())
				err = rows.Scan(&id, &name, &profileName)
				assert.NoError(t, err)
				assert.Equal(t, 1, id)
				assert.Equal(t, "zakirkun", name)
				assert.Equal(t, "Developer", profileName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setup(mock)

			// Create builder and execute raw query
			builder := New()
			query := builder.Raw(tt.sql, tt.args...).WithDB(db)

			// Validate the query
			tt.validate(t, query)

			// Verify all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			t.Log("---- Pass ----")
		})
	}
}

func TestRawQueryWithTransaction(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Setup transaction expectations
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").
		WithArgs("zakirkun", "zakir@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE profiles").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create builder
	builder := New()

	// Execute first query
	_, err = builder.Raw("INSERT INTO users(name, email) VALUES($1, $2)",
		"zakirkun", "zakir@example.com").
		WithTx(tx).
		Exec()
	assert.NoError(t, err)

	// Execute second query
	_, err = builder.Raw("UPDATE profiles SET verified = true WHERE user_id = $1", 1).
		WithTx(tx).
		Exec()
	assert.NoError(t, err)

	// Commit transaction
	err = tx.Commit()
	assert.NoError(t, err)

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	t.Log("---- Pass ----")
}

func TestRawQueryErrors(t *testing.T) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Test query error
	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	builder := New()
	rows, err := builder.Raw("SELECT * FROM users").WithDB(db).Query()
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.Nil(t, rows)

	// Test exec error
	mock.ExpectExec("INSERT").WillReturnError(sql.ErrConnDone)

	result, err := builder.Raw("INSERT INTO users(name) VALUES($1)", "zakirkun").
		WithDB(db).
		Exec()
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.Nil(t, result)

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	t.Log("---- Pass ----")
}
