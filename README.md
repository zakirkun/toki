# Toki SQL Query Builder

[![Go Reference](https://pkg.go.dev/badge/github.com/zakirkun/toki.svg)](https://pkg.go.dev/github.com/zakirkun/toki)
[![Go Report Card](https://goreportcard.com/badge/github.com/zakirkun/toki)](https://goreportcard.com/report/github.com/zakirkun/toki)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


Toki is a fast and efficient SQL query builder for Go that helps you build SQL statements dynamically at runtime. It focuses on performance, safety, and ease of use.

## Features

- 🚀 High-performance query building
- 🔒 Automatic placeholder conversion (? to $1, $2, ...)
- 💾 Memory pooling for better performance
- 📦 Transaction support
- 🎯 Dynamic query construction
- 🛡️ SQL injection prevention
- 📊 Structure binding support
- ⚡ Memory efficient operations
- 🔍 Expression support

## Installation

```bash
go get github.com/zakirkun/toki
```

## Quick Start

```go
package main

import (
    "database/sql"
    "log"

    "github.com/zakirkun/toki"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:password@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create a new builder
    builder := toki.New()

    // Build a SELECT query
    query := builder.
        Select("id", "name", "email").
        From("users").
        Where("age > ?", 18).
        OrderBy("created_at DESC")

    // Prepare and execute
    stmt, err := query.Prepare(db)
    if err != nil {
        log.Fatal(err)
    }

    rows, err := stmt.Query()
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
}
```

## Core Features

### Query Building

```go
// SELECT query
builder.
    Select("id", "name").
    From("users").
    Where("age > ?", 18).
    AndWhere("status = ?", "active").
    OrderBy("created_at DESC")

// INSERT query
builder.
    Insert("users", "name", "email").
    Values("John Doe", "john@example.com")

// UPDATE query
builder.
    Update("users").
    Set(map[string]interface{}{
        "name": "Jane Doe",
        "updated_at": "NOW()",
    }).
    Where("id = ?", 1)

// DELETE query
builder.
    Delete("users").
    Where("status = ?", "inactive")
```

### Transaction Support

```go
// Start a transaction
tx, err := toki.Begin(db)
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback() // Rollback if not committed

// Use transaction in builder
builder := toki.New().WithTransaction(tx)

// Execute queries
stmt, err := builder.
    Insert("users").
    Values("John", "john@example.com").
    Prepare(db)

if err != nil {
    log.Fatal(err)
}

// Commit transaction
if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

### Structure Binding

```go
type User struct {
    ID        int       `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
}

// Bind structure to query
user := User{
    Name:  "John Doe",
    Email: "john@example.com",
}

bindings := builder.Bind(&user)
```

### SQL Expressions

```go
// Using raw SQL expressions
builder.
    Update("counters").
    Set(map[string]interface{}{
        "counter": toki.Raw("counter + 1"),
        "updated_at": toki.Raw("NOW()"),
    }).
    Where("id = ?", 1)
```

## Performance Features

### Memory Pooling
Toki uses sync.Pool to reuse allocated memory for string building, which significantly reduces garbage collection pressure under high load.

```go
// Memory pooling is automatically handled
builder := toki.New() // Creates a builder with memory pool
```

### Placeholder Conversion
Automatic and efficient conversion of SQL placeholders:

```go
// Write queries with ? placeholders
query := builder.
    Select("*").
    From("users").
    Where("age > ?", 18).
    AndWhere("status = ?", "active")

// Automatically converts to $1, $2, etc. for PostgreSQL
// SELECT * FROM users WHERE age > $1 AND status = $2
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you have any questions or need help, please open an issue in the GitHub repository.

