# Fayl SDK

[![Go Version](https://img.shields.io/badge/Go-1.24.2+-blue.svg)](https://golang.org)

**Fayl** is a Go library that simplifies database operations through SQL templates and type-safe result mapping. It provides a clean abstraction over raw SQL while maintaining full control over your queries.

## 🚀 Features

- **SQL-First Design**: Write and organize your SQL queries in separate files
- **Template-Based Queries**: Use Go templates for dynamic SQL generation
- **Type-Safe Scanning**: Automatic mapping between SQL results and Go structs
- **Transaction Support**: Built-in transaction management with automatic rollback
- **Pagination**: Support for both offset and cursor-based pagination
- **Parameter Binding**: Safe parameter substitution with SQL injection protection
- **Flexible Scanning**: Support for maps, structs, and custom writers
- **Database Agnostic**: Works with any database supported by Go's `database/sql`

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Usage Examples](#usage-examples)
- [API Reference](#api-reference)
- [Configuration](#configuration)

## 📦 Installation

```bash
go get github.com/redhajuanda/fayl
```

## 🏃‍♂️ Quick Start

### 1. Initialize the Client

```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/redhajuanda/fayl"
    _ "github.com/go-sql-driver/mysql" // MariaDB driver
)

func main() {
    // Open database connection
    db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname?parseTime=true")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Initialize Fayl client
    client, err := fayl.Init(logger, fayl.Option{
        DB:            db,
        QueryLocation: "./queries", // Base directory for SQL query files (base path excluded from query names)
        DriverName:    "mysql", // Must match the driver used in sql.Open()
        Placeholder:   fayl.Question, // Use placeholder `?` for MariaDB
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2. Create SQL Query Files

Create a directory structure for your queries:

```
queries/
├── user/
│   ├── GetUser.sql
│   ├── CreateUser.sql
│   └── ListUsers.sql
└── product/
    ├── GetProduct.sql
    └── SearchProducts.sql
```

Example SQL file (`queries/user/GetUser.sql`):
```sql
SELECT id, name, email, created_at 
FROM users 
WHERE id = {{ .id }}
```

### 3. Execute Queries

```go
// Define your struct
type User struct {
    ID        int       `fayl:"id"`
    Name      string    `fayl:"name"`
    Email     string    `fayl:"email"`
    CreatedAt time.Time `fayl:"created_at"`
}

// Get a single user
var user User
err := client.Run("user.GetUser").
    WithParam("id", 1).
    ScanStruct(&user).
    Query(context.Background())

// Get multiple users
var users []User
err = client.Run("user.ListUsers").
    ScanStructs(&users).
    Query(context.Background())
```

## 🧠 Core Concepts

### Query Organization

Fayl organizes queries using a dot notation system based on your file structure:

- File: `queries/user/GetUser.sql` → Query name: `user.GetUser`
- File: `queries/product/SearchProducts.sql` → Query name: `product.SearchProducts`

### Parameter Binding

Fayl uses Go templates for parameter substitution:

```sql
-- queries/user/SearchUsers.sql
SELECT id, name, email 
FROM users 
WHERE name ILIKE {{ .name }} 
AND is_active = {{ .is_active }}
```

### Scanning Results

Fayl provides multiple ways to scan query results:

- **Structs**: `ScanStruct(&user)` / `ScanStructs(&users)`
- **Maps**: `ScanMap(map[string]any{})` / `ScanMaps(&[]map[string]any{})`
- **Custom Writers**: `ScanWriter(io.Writer)`

## 📚 Usage Examples

### Basic CRUD Operations

```go
// Create a user
type CreateUserParams struct {
    Name  string `fayl:"name"`
    Email string `fayl:"email"`
}

params := CreateUserParams{
    Name:  "John Doe",
    Email: "john@example.com",
}

result, err := client.Run("user.CreateUser").
    WithParams(params).
    Exec(context.Background())

// Update a user
updateParams := map[string]any{
    "id":    1,
    "name":  "Jane Doe",
    "email": "jane@example.com",
}

_, err = client.Run("user.UpdateUser").
    WithParams(updateParams).
    Exec(context.Background())

// Delete a user
_, err = client.Run("user.DeleteUser").
    WithParam("id", 1).
    Exec(context.Background())
```

### Pagination

```go
import "github.com/redhajuanda/perkakas/pagination"

// Offset-based pagination
offsetPagination := &pagination.Pagination{
    Type:    "offset",
    Page:    1,
    PerPage: 10,
}

var users []User
err := client.Run("user.ListUsers").
    WithPagination(offsetPagination).
    ScanStructs(&users).
    Query(context.Background())

// Cursor-based pagination
cursorPagination := &pagination.Pagination{
    Type:      "cursor",
    Cursor:    "eyJpZCI6MTAwfQ==",
    PerPage:   10,
}
err = client.Run("user.ListUsers").
    WithPagination(cursorPagination).
    WithOrderBy("id"). // Order by is required when using cursor pagination. The column name must match a struct field tag in the result, as this column's value is used as the cursor pointer to track the last query position
    ScanStructs(&users).
    Query(context.Background())
```

### Cursor Pagination Details

When using cursor-based pagination, there are several important considerations to ensure proper functionality:

#### Required Order By
Cursor pagination **requires** an `OrderBy` clause. The column name must match a struct field tag in the result, as this column's value is used as the cursor pointer to track the last query position.

#### Unique Order Combination
The combination of columns in the `OrderBy` clause must be **guaranteed unique**. If not, the pagination will become inconsistent and may skip or duplicate records.

#### Handling Nullable Columns
If the order column is nullable, you need to specify `nullable` after the column name. For example: `name nullable` or `-name nullable`. Note that only one column can be marked as nullable in the OrderBy clause because cursor pagination requires a deterministic ordering, and having multiple nullable columns would create ambiguity in the cursor position tracking. Since nullable columns are definitely not unique, you must combine them with another column that is unique to ensure the ordering combination remains unique.


### Transactions

```go
type User struct {
    ID    int    `fayl:"id"`
    Name  string `fayl:"name"`
    Email string `fayl:"email"`
}

type Order struct {
    ID     int    `fayl:"id"`
    UserID int    `fayl:"user_id"`
    Amount int    `fayl:"amount"`
}

// Create user and order in a transaction
result, err := client.WithTransaction(context.Background(), func(ctx context.Context, tx *fayl.Tx) (any, error) {
    // Create user
    userParams := map[string]any{
        "name":  "John Doe",
        "email": "john@example.com",
    }
    
    userResult, err := tx.Run("user.CreateUser").WithParams(userParams).Exec(ctx)
    if err != nil {
        return nil, err
    }
    
    userID := userResult.LastInsertID
    
    // Create order
    orderParams := map[string]any{
        "user_id": userID,
        "amount":  100,
    }
    
    _, err = tx.Run("order.CreateOrder").WithParams(orderParams).Exec(ctx)
    if err != nil {
        return nil, err
    }
    
    return userID, nil
})
```

### Complex Queries with Conditions

```sql
-- queries/user/SearchUsers.sql
SELECT id, name, email, created_at
FROM users
WHERE 1=1
  {{ if .name }}AND name ILIKE {{ .name }}{{ end }}
  {{ if .email }}AND email ILIKE {{ .email }}{{ end }}
  {{ if .is_active }}AND is_active = {{ .is_active }}{{ end }}
  {{ if .created_after }}AND created_at >= {{ .created_after }}{{ end }}
ORDER BY created_at DESC
{{ if .limit }}LIMIT {{ .limit }}{{ end }}
```

```go
// Execute with conditional parameters
params := map[string]any{
    "name":          "%john%",
    "is_active":     true,
    "created_after": time.Now().AddDate(0, -1, 0), // Last month
    "limit":         50,
}

var users []User
err := client.Run("user.SearchUsers").
    WithParams(params).
    ScanStructs(&users).
    Query(context.Background())
```

### Working with JSON Data

```sql
-- queries/user/UpdateUserPreferences.sql
UPDATE users 
SET preferences = {{ .preferences }}
WHERE id = {{ .id }}
```

```go
type UserPreferences struct {
    Theme     string   `json:"theme"`
    Language  string   `json:"language"`
    Notifications []string `json:"notifications"`
}

preferences := UserPreferences{
    Theme:     "dark",
    Language:  "en",
    Notifications: []string{"email", "push"},
}

params := map[string]any{
    "id":          1,
    "preferences": preferences,
}

_, err := client.Run("user.UpdateUserPreferences").
    WithParams(params).
    Exec(context.Background())
```

## 🔧 Configuration

### Client Options

```go
type Option struct {
    DB            *sql.DB           // Database connection
    QueryLocation string            // Path to SQL files directory
    DriverName    string            // Database driver name
    Placeholder   parser.Placeholder // Placeholder format
}
```

### Supported Placeholders

Fayl provides several placeholder formats to support different database systems:

- **`fayl.Dollar`** (uses `$1`, `$2`, etc.)
  - PostgreSQL

- **`fayl.Question`** (uses `?`)
  - MySQL
  - MariaDB
  - SQLite
  - Snowflake

- **`fayl.Colon`** (uses `:1`, `:2`, etc.)
  - Oracle

- **`fayl.AtP`** (uses `@p1`, `@p2`, etc.)
  - SQL Server

### Struct Tags

Use the `fayl` tag to map database columns to struct fields:

```go
type User struct {
    ID        int       `fayl:"id"`
    Name      string    `fayl:"name"`
    Email     string    `fayl:"email"`
    CreatedAt time.Time `fayl:"created_at"`
    UpdatedAt time.Time `fayl:"updated_at"`
}
```

## 📖 API Reference

### Client Methods

- `Run(queryName string) Runnerer` - Start a new query execution
- `WithTransaction(ctx context.Context, callback TxFunc) (any, error)` - Execute in transaction

### Runner Methods

- `WithParam(key string, value any) Runnerer` - Add single parameter
- `WithParams(params any) Runnerer` - Add multiple parameters
- `WithPagination(pagination *pagination.Pagination) Runnerer` - Add pagination
- `WithOrderBy(orderBy ...string) Runnerer` - Add ordering
- `ScanStruct(dest any) Runnerer` - Scan to single struct
- `ScanStructs(dest any) Runnerer` - Scan to slice of structs
- `ScanMap(dest map[string]any) Runnerer` - Scan to map
- `ScanMaps(dest *[]map[string]any) Runnerer` - Scan to slice of maps
- `ScanWriter(dest io.Writer) Runnerer` - Scan to writer
- `Exec(ctx context.Context) (*ResultExec, error)` - Execute without scanning
- `Query(ctx context.Context) error` - Execute and scan
