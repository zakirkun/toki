package toki

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Builder represents the main query builder structure
type Builder struct {
	parts    []string
	args     []interface{}
	argIndex int
	pool     *sync.Pool
	table    string
	tx       *Transaction
}

// New creates a new query builder
func New() *Builder {
	return &Builder{
		pool: &sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
	}
}

// WithTransaction sets the transaction for the builder
func (b *Builder) WithTransaction(tx *Transaction) *Builder {
	b.tx = tx
	return b
}

// Select initializes a SELECT query
func (b *Builder) Select(columns ...string) *Builder {
	b.parts = append(b.parts, fmt.Sprintf("SELECT %s", strings.Join(columns, ", ")))
	return b
}

// From adds FROM clause
func (b *Builder) From(table string) *Builder {
	b.table = table
	b.parts = append(b.parts, fmt.Sprintf("FROM %s", b.table))
	return b
}

// Where adds WHERE conditions
func (b *Builder) Where(condition string, args ...interface{}) *Builder {
	if len(b.parts) > 0 && !strings.HasSuffix(b.parts[len(b.parts)-1], "WHERE") {
		b.parts = append(b.parts, "WHERE")
	}
	b.parts = append(b.parts, b.convertPlaceholders(condition))
	b.args = append(b.args, args...)
	return b
}

// AndWhere adds AND condition
func (b *Builder) AndWhere(condition string, args ...interface{}) *Builder {
	b.parts = append(b.parts, "AND", b.convertPlaceholders(condition))
	b.args = append(b.args, args...)
	return b
}

// OrWhere adds OR condition
func (b *Builder) OrWhere(condition string, args ...interface{}) *Builder {
	b.parts = append(b.parts, "OR", b.convertPlaceholders(condition))
	b.args = append(b.args, args...)
	return b
}

// OrderBy adds ORDER BY clause
func (b *Builder) OrderBy(columns ...string) *Builder {
	b.parts = append(b.parts, fmt.Sprintf("ORDER BY %s", strings.Join(columns, ", ")))
	return b
}

// Update initializes an UPDATE query
func (b *Builder) Update(table string) *Builder {
	b.parts = append(b.parts, fmt.Sprintf("UPDATE %s", table))
	return b
}

// Set adds SET clause for UPDATE
func (b *Builder) Set(updates map[string]interface{}) *Builder {

	sets := make([]string, 0, len(updates))
	for col, val := range updates {
		if expr, ok := val.(SQLExpression); ok {
			sets = append(sets, fmt.Sprintf("%s = %s", col, expr.SQL()))
		} else {
			b.argIndex++
			sets = append(sets, fmt.Sprintf("%s = $%d", col, b.argIndex))
			b.args = append(b.args, val)
		}
	}

	b.parts = append(b.parts, fmt.Sprintf("SET %s", strings.Join(sets, ", ")))
	return b
}

// Insert initializes an INSERT query
func (b *Builder) Insert(table string, columns ...string) *Builder {
	b.parts = append(b.parts, fmt.Sprintf("INSERT INTO %s (%s)", table, strings.Join(columns, ", ")))

	return b
}

// Values adds VALUES clause for INSERT
func (b *Builder) Values(values ...interface{}) *Builder {
	placeholders := make([]string, len(values))
	for i := range values {
		b.argIndex++
		placeholders[i] = fmt.Sprintf("$%d", b.argIndex)
	}

	b.parts = append(b.parts, fmt.Sprintf("VALUES (%s)", strings.Join(placeholders, ", ")))
	b.args = append(b.args, values...)
	return b
}

// Delete initializes a DELETE query
func (b *Builder) Delete(table string) *Builder {
	b.parts = append(b.parts, fmt.Sprintf("DELETE FROM %s", table))
	return b
}

// DeleteFrom is an alias for Delete for more expressive API
func (b *Builder) DeleteFrom(table string) *Builder {
	return b.Delete(table)
}

// Returning adds a RETURNING clause to the DELETE statement
func (b *Builder) Returning(columns ...string) *Builder {
	if len(columns) > 0 {
		b.parts = append(b.parts, "RETURNING", strings.Join(columns, ", "))
	}
	return b
}

// String builds the final query string
func (b *Builder) String() string {
	sb := b.pool.Get().(*strings.Builder)
	defer func() {
		sb.Reset()
		b.pool.Put(sb)
	}()

	for i, part := range b.parts {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(part)
	}

	return sb.String()
}

// Bind creates a struct binding for database columns
func (b *Builder) Bind(dest interface{}) map[string]interface{} {
	val := reflect.ValueOf(dest)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	result := make(map[string]interface{})

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("db")
		if tag != "" {
			result[tag] = val.Field(i).Interface()
		}
	}

	if b.table == "" {
		b.table = strings.ToLower(typ.Name())
	}

	return result
}

// convertPlaceholders converts ? placeholders to $1, $2, etc.
func (b *Builder) convertPlaceholders(query string) string {
	result := strings.Builder{}

	for _, c := range query {
		if c == '?' {
			b.argIndex++
			result.WriteString(fmt.Sprintf("$%d", b.argIndex))
		} else {
			result.WriteRune(c)
		}
	}

	return result.String()
}
