package toki

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestTime is a fixed time for testing
var (
	TestTime = time.Date(2024, 12, 23, 5, 45, 29, 0, time.UTC)
	TestUser = "zakirkun"
)

// Mock interfaces and structs for testing
func setupTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Builder) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}

	builder := New()

	return db, mock, builder
}

func TestSelect(t *testing.T) {

	tests := []struct {
		name     string
		build    func(*Builder) *Builder
		expected string
		args     []interface{}
	}{
		{
			name: "Simple select",
			build: func(b *Builder) *Builder {
				return b.Select("id", "name").From("users")
			},
			expected: "SELECT id, name FROM users",
			args:     nil,
		},
		{
			name: "Select with where clause",
			build: func(b *Builder) *Builder {
				return b.Select("*").From("users").Where("id = ?", 1)
			},
			expected: "SELECT * FROM users WHERE id = $1",
			args:     []interface{}{1},
		},
		{
			name: "Select with multiple conditions",
			build: func(b *Builder) *Builder {
				return b.Select("*").
					From("users").
					Where("age > ?", 18).
					AndWhere("status = ?", "active").
					OrderBy("created_at DESC")
			},
			expected: "SELECT * FROM users WHERE age > $1 AND status = $2 ORDER BY created_at DESC",
			args:     []interface{}{18, "active"},
		},
	}

	runBuilderTests(t, tests)
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*Builder) *Builder
		expected string
		args     []interface{}
	}{
		{
			name: "Simple insert",
			build: func(b *Builder) *Builder {
				return b.Insert("users", "name", "email").
					Values("zakirkun", "zakir@example.com")
			},
			expected: "INSERT INTO users (name, email) VALUES ($1, $2)",
			args:     []interface{}{"zakirkun", "zakir@example.com"},
		},
		{
			name: "Insert with returning",
			build: func(b *Builder) *Builder {
				return b.Insert("users", "name", "email").
					Values("zakirkun", "zakir@example.com").
					Returning("id")
			},
			expected: "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
			args:     []interface{}{"zakirkun", "zakir@example.com"},
		},
	}

	runBuilderTests(t, tests)
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*Builder) *Builder
		expected string
		args     []interface{}
	}{
		{
			name: "Simple update",
			build: func(b *Builder) *Builder {
				return b.Update("users").
					Set(map[string]interface{}{
						"name":       "New Name",
						"updated_at": TestTime,
					}).
					Where("id = ?", 1)
			},
			expected: "UPDATE users SET name = $1, updated_at = $2 WHERE id = $3",
			args:     []interface{}{"New Name", TestTime, 1},
		},
		{
			name: "Update with expression",
			build: func(b *Builder) *Builder {
				return b.Update("counters").
					Set(map[string]interface{}{
						"count": Raw("count + 1"),
					}).
					Where("id = ?", 1)
			},
			expected: "UPDATE counters SET count = count + 1 WHERE id = $1",
			args:     []interface{}{1},
		},
	}

	runBuilderTests(t, tests)
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*Builder) *Builder
		expected string
		args     []interface{}
	}{
		{
			name: "Simple delete",
			build: func(b *Builder) *Builder {
				return b.Delete("users").Where("id = ?", 1)
			},
			expected: "DELETE FROM users WHERE id = $1",
			args:     []interface{}{1},
		},
		{
			name: "Delete with multiple conditions",
			build: func(b *Builder) *Builder {
				return b.Delete("users").
					Where("status = ?", "inactive").
					AndWhere("last_login < ?", TestTime)
			},
			expected: "DELETE FROM users WHERE status = $1 AND last_login < $2",
			args:     []interface{}{"inactive", TestTime},
		},
	}

	runBuilderTests(t, tests)
}

func TestTransaction(t *testing.T) {
	db, mock, builder := setupTest(t)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users .*").
		WithArgs(TestUser, "zakir@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := Begin(db)
	assert.NoError(t, err)

	stmt, err := builder.
		WithTransaction(tx).
		Insert("users", "name", "email").
		Values(TestUser, "zakir@example.com").
		Prepare(db)

	assert.NoError(t, err)

	_, err = stmt.Exec()
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())

	t.Log("---- Pass ----")
}

func TestStructBinding(t *testing.T) {
	type User struct {
		ID        int       `db:"id"`
		Name      string    `db:"name"`
		Email     string    `db:"email"`
		CreatedAt time.Time `db:"created_at"`
	}

	user := User{
		ID:        0,
		Name:      "zakirkun",
		Email:     "zakir@example.com",
		CreatedAt: TestTime,
	}

	builder := New()
	bindings := builder.Bind(&user)

	expected := map[string]interface{}{
		"id":         0,
		"name":       "zakirkun",
		"email":      "zakir@example.com",
		"created_at": TestTime,
	}

	if !reflect.DeepEqual(bindings, expected) {
		t.Errorf("Struct binding failed.\nExpected: %v\nGot: %v", expected, bindings)
	}

	t.Log("---- Pass ----")
}

// Test Raw SQL expressions
func TestRawExpression(t *testing.T) {
	expr := Raw("NOW()")
	if expr.SQL() != "NOW()" {
		t.Errorf("Raw expression mismatch.\nExpected: NOW()\nGot: %s", expr.SQL())
	}

	t.Log("---- Pass ----")
}

// Test placeholder conversion
func TestPlaceholderConversion(t *testing.T) {
	builder := New()
	query := builder.convertPlaceholders("SELECT * FROM users WHERE id = ? AND name = ?")
	expected := "SELECT * FROM users WHERE id = $1 AND name = $2"

	if query != expected {
		t.Errorf("Placeholder conversion failed.\nExpected: %s\nGot: %s", expected, query)
	}

	t.Log("---- Pass ----")
}

// Helper function to run builder tests
func runBuilderTests(t *testing.T, tests []struct {
	name     string
	build    func(*Builder) *Builder
	expected string
	args     []interface{}
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := New()
			result := tt.build(builder)
			query := result.String()

			// Remove extra spaces for comparison
			query = strings.TrimSpace(query)
			expected := strings.TrimSpace(tt.expected)

			if query != expected {
				t.Errorf("Query mismatch.\nExpected: %s\nGot: %s", expected, query)
			}

			if tt.args != nil {
				if len(result.args) != len(tt.args) {
					t.Errorf("Arguments length mismatch.\nExpected: %d\nGot: %d", len(tt.args), len(result.args))
					return
				}

				for i, arg := range tt.args {
					if !reflect.DeepEqual(result.args[i], arg) {
						t.Errorf("Argument at position %d mismatch.\nExpected: %v\nGot: %v", i, arg, result.args[i])
					}
				}
			}

			t.Log("Query:", query)
			t.Log("Args:", result.args)
			t.Log("---- Pass ----")
		})
	}
}

// Helper function to scan rows into map
func scanRows(rows *sql.Rows) []map[string]interface{} {
	var results []map[string]interface{}
	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)
		entry := make(map[string]interface{})

		for i, col := range columns {
			val := values[i]
			entry[col] = val
		}

		results = append(results, entry)
	}

	return results
}
