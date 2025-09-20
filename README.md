# GraphStructManager - Gremlin Query Builder

A type-safe, chainable query builder for Gremlin graph databases in Go. This ORM provides an intuitive interface for building and executing Gremlin queries with full type safety.

## Table of Contents

- [Overview](#overview)
- [Setup](#setup)
- [Query Builder Functions](#query-builder-functions)
- [Complete Examples](#complete-examples)
- [Comparison Operators](#comparison-operators)

## Overview

The query builder uses Go generics to provide type-safe operations on vertex types that implement the `VertexType` interface. All functions are chainable, allowing for fluent query construction.

## Setup

First, define your vertex struct with the required gremlin tags:

```go
type TestVertex struct {
    types.Vertex                               // Anonymous embedding required
    Name        string   `gremlin:"name"`      // Field with gremlin tag
    Age         int      `gremlin:"age"`
    Email       string   `gremlin:"email"`
    Tags        []string `gremlin:"tags"`
}
```

Connect to your Gremlin database:

```go
db, err := GSM.Open("ws://localhost:8182")
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

## Query Builder Functions

### NewQuery[T]

Creates a new query builder for the specified vertex type.

**Signature:**
```go
func NewQuery[T VertexType](db *GremlinDriver) *Query[T]
```

**Usage:**
```go
// Create a new query builder for TestVertex
query := GSM.NewQuery[TestVertex](db)

// Or use the convenience function
query := GSM.Model[TestVertex](db)
```

### Where

Adds a condition to the query using comparison operators.

**Signature:**
```go
func (q *Query[T]) Where(field string, operator comparator.Comparator, value any) *Query[T]
```

**Examples:**
```go
// Equal comparison
users := GSM.Model[TestVertex](db).Where("name", comparator.EQ, "John")

// Not equal
users := GSM.Model[TestVertex](db).Where("age", comparator.NEQ, 25)

// Greater than
users := GSM.Model[TestVertex](db).Where("age", comparator.GT, 18)

// Greater than or equal
users := GSM.Model[TestVertex](db).Where("age", comparator.GTE, 21)

// Less than
users := GSM.Model[TestVertex](db).Where("age", comparator.LT, 65)

// Less than or equal
users := GSM.Model[TestVertex](db).Where("age", comparator.LTE, 30)

// In array
users := GSM.Model[TestVertex](db).Where("name", comparator.IN, []any{"John", "Jane", "Bob"})

// Contains (for string fields)
users := GSM.Model[TestVertex](db).Where("email", comparator.CONTAINS, "@gmail.com")

// Chain multiple conditions
users := GSM.Model[TestVertex](db).
    Where("age", comparator.GT, 18).
    Where("email", comparator.CONTAINS, "@company.com")
```

### WhereTraversal

Adds a custom Gremlin traversal condition for advanced queries.

**Signature:**
```go
func (q *Query[T]) WhereTraversal(traversal *gremlingo.GraphTraversal) *Query[T]
```

**Examples:**
```go
// Custom traversal with has step
users := GSM.Model[TestVertex](db).
    WhereTraversal(gremlingo.T__.Has("name", "John"))

// Complex traversal
users := GSM.Model[TestVertex](db).
    WhereTraversal(gremlingo.T__.Has("age", gremlingo.P.Between(25, 35)))

// Combine with regular Where conditions
users := GSM.Model[TestVertex](db).
    Where("name", comparator.EQ, "John").
    WhereTraversal(gremlingo.T__.Has("email", gremlingo.P.StartingWith("j")))
```

### Dedup

Removes duplicate results from the query.

**Signature:**
```go
func (q *Query[T]) Dedup() *Query[T]
```

**Examples:**
```go
// Remove duplicates
uniqueUsers := GSM.Model[TestVertex](db).
    Where("tags", comparator.CONTAINS, "developer").
    Dedup()

// Chain with other operations
users := GSM.Model[TestVertex](db).
    Where("age", comparator.GT, 25).
    Dedup().
    OrderBy("name")
```

### Limit

Sets the maximum number of results to return.

**Signature:**
```go
func (q *Query[T]) Limit(limit int) *Query[T]
```

**Examples:**
```go
// Get first 10 users
users := GSM.Model[TestVertex](db).
    OrderBy("name").
    Limit(10)

// Top 5 oldest users
oldestUsers := GSM.Model[TestVertex](db).
    OrderByDesc("age").
    Limit(5)

// Combine with where conditions
activeUsers := GSM.Model[TestVertex](db).
    Where("status", comparator.EQ, "active").
    Limit(20)
```

### Offset

Sets the number of results to skip (for pagination).

**Signature:**
```go
func (q *Query[T]) Offset(offset int) *Query[T]
```

**Examples:**
```go
// Skip first 20 results (page 2 with 20 per page)
users := GSM.Model[TestVertex](db).
    OrderBy("name").
    Offset(20).
    Limit(20)

// Get results 50-100
users := GSM.Model[TestVertex](db).
    Offset(50).
    Limit(50)

// Pagination helper function
func getPage(db *GSM.GremlinDriver, page, pageSize int) ([]TestVertex, error) {
    return GSM.Model[TestVertex](db).
        OrderBy("id").
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find()
}
```

### OrderBy

Adds ascending ordering to the query.

**Signature:**
```go
func (q *Query[T]) OrderBy(field string) *Query[T]
```

**Examples:**
```go
// Order by name (ascending)
users := GSM.Model[TestVertex](db).
    OrderBy("name")

// Order by age, then by name
users := GSM.Model[TestVertex](db).
    OrderBy("age").
    OrderBy("name")

// Combine with filtering
youngUsers := GSM.Model[TestVertex](db).
    Where("age", comparator.LT, 30).
    OrderBy("age")
```

### OrderByDesc

Adds descending ordering to the query.

**Signature:**
```go
func (q *Query[T]) OrderByDesc(field string) *Query[T]
```

**Examples:**
```go
// Order by age (descending)
users := GSM.Model[TestVertex](db).
    OrderByDesc("age")

// Multiple ordering: newest first, then by name
users := GSM.Model[TestVertex](db).
    OrderByDesc("lastModified").
    OrderBy("name")

// Top earners
topUsers := GSM.Model[TestVertex](db).
    OrderByDesc("salary").
    Limit(10)
```

### Find

Executes the query and returns all matching results.

**Signature:**
```go
func (q *Query[T]) Find() ([]T, error)
```

**Examples:**
```go
// Get all users
allUsers, err := GSM.Model[TestVertex](db).Find()
if err != nil {
    return err
}

// Get filtered results
activeUsers, err := GSM.Model[TestVertex](db).
    Where("status", comparator.EQ, "active").
    Find()

// Get paginated results
users, err := GSM.Model[TestVertex](db).
    OrderBy("name").
    Limit(50).
    Find()

// Complex query
developers, err := GSM.Model[TestVertex](db).
    Where("department", comparator.EQ, "engineering").
    Where("experience", comparator.GTE, 2).
    OrderByDesc("salary").
    Find()
```

### First

Executes the query and returns the first result.

**Signature:**
```go
func (q *Query[T]) First() (T, error)
```

**Examples:**
```go
// Get first user by name
user, err := GSM.Model[TestVertex](db).
    Where("name", comparator.EQ, "John").
    First()
if err != nil {
    return err
}

// Get oldest user
oldestUser, err := GSM.Model[TestVertex](db).
    OrderByDesc("age").
    First()

// Get user with specific email
user, err := GSM.Model[TestVertex](db).
    Where("email", comparator.EQ, "john@example.com").
    First()

// Handle not found
user, err := GSM.Model[TestVertex](db).
    Where("id", comparator.EQ, nonExistentId).
    First()
if err != nil {
    if err.Error() == "no more results" {
        // Handle not found case
        fmt.Println("User not found")
    } else {
        // Handle other errors
        return err
    }
}
```

### Count

Returns the number of matching results without retrieving the actual data.

**Signature:**
```go
func (q *Query[T]) Count() (int, error)
```

**Examples:**
```go
// Count all users
totalUsers, err := GSM.Model[TestVertex](db).Count()
if err != nil {
    return err
}

// Count active users
activeCount, err := GSM.Model[TestVertex](db).
    Where("status", comparator.EQ, "active").
    Count()

// Count users in age range
adultsCount, err := GSM.Model[TestVertex](db).
    Where("age", comparator.GTE, 18).
    Where("age", comparator.LTE, 65).
    Count()

// Check if any users exist with condition
hasAdmins, err := GSM.Model[TestVertex](db).
    Where("role", comparator.EQ, "admin").
    Count()
if err != nil {
    return err
}
if hasAdmins > 0 {
    fmt.Println("Admin users exist")
}
```

### Delete

Deletes all vertices matching the query conditions.

**Signature:**
```go
func (q *Query[T]) Delete() error
```

**Examples:**
```go
// Delete specific user
err := GSM.Model[TestVertex](db).
    Where("email", comparator.EQ, "user@example.com").
    Delete()

// Delete inactive users
err := GSM.Model[TestVertex](db).
    Where("status", comparator.EQ, "inactive").
    Delete()

// Delete users older than 100 (cleanup)
err := GSM.Model[TestVertex](db).
    Where("age", comparator.GT, 100).
    Delete()

// Delete with multiple conditions
err := GSM.Model[TestVertex](db).
    Where("department", comparator.EQ, "temp").
    Where("lastLogin", comparator.LT, oneYearAgo).
    Delete()

if err != nil {
    log.Printf("Failed to delete users: %v", err)
    return err
}
```

## Complete Examples

### Basic CRUD Operations

```go
func main() {
    // Setup
    db, err := GSM.Open("ws://localhost:8182")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create a user
    newUser := TestVertex{
        Name:  "Alice Johnson",
        Age:   28,
        Email: "alice@example.com",
        Tags:  []string{"developer", "golang", "senior"},
    }

    err = GSM.Create(db, &newUser)
    if err != nil {
        log.Fatal(err)
    }

    // Read - Find user by email
    user, err := GSM.Model[TestVertex](db).
        Where("email", comparator.EQ, "alice@example.com").
        First()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found user: %+v\n", user)

    // Update would typically involve Create with existing ID

    // Delete - Remove user
    err = GSM.Model[TestVertex](db).
        Where("email", comparator.EQ, "alice@example.com").
        Delete()
    if err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Querying

```go
func advancedQueries(db *GSM.GremlinDriver) {
    // Pagination
    page := 2
    pageSize := 10
    users, err := GSM.Model[TestVertex](db).
        OrderBy("name").
        Offset((page-1) * pageSize).
        Limit(pageSize).
        Find()

    // Search with multiple filters
    seniorDevelopers, err := GSM.Model[TestVertex](db).
        Where("age", comparator.GTE, 25).
        Where("experience", comparator.GT, 3).
        Where("tags", comparator.CONTAINS, "senior").
        OrderByDesc("experience").
        Find()

    // Count and statistics
    totalDevelopers, err := GSM.Model[TestVertex](db).
        Where("tags", comparator.CONTAINS, "developer").
        Count()

    juniorCount, err := GSM.Model[TestVertex](db).
        Where("tags", comparator.CONTAINS, "junior").
        Count()

    fmt.Printf("Total developers: %d, Junior: %d\n", totalDevelopers, juniorCount)

    // Complex query with custom traversal
    complexResults, err := GSM.Model[TestVertex](db).
        Where("department", comparator.EQ, "engineering").
        WhereTraversal(gremlingo.T__.Has("salary", gremlingo.P.Between(50000, 100000))).
        OrderByDesc("lastModified").
        Limit(20).
        Find()
}
```

### Error Handling Patterns

```go
func handleQueryErrors(db *GSM.GremlinDriver) {
    // Handle "not found" gracefully
    user, err := GSM.Model[TestVertex](db).
        Where("id", comparator.EQ, "non-existent-id").
        First()

    if err != nil {
        if strings.Contains(err.Error(), "no more results") {
            fmt.Println("User not found")
            // Handle not found case
            return
        }
        // Handle other errors
        log.Printf("Query error: %v", err)
        return
    }

    // Check if results exist before processing
    count, err := GSM.Model[TestVertex](db).
        Where("status", comparator.EQ, "pending").
        Count()

    if err != nil {
        log.Printf("Count error: %v", err)
        return
    }

    if count == 0 {
        fmt.Println("No pending users found")
        return
    }

    // Process pending users
    pendingUsers, err := GSM.Model[TestVertex](db).
        Where("status", comparator.EQ, "pending").
        Find()
    // ... process users
}
```

## Comparison Operators

The following comparison operators are available in the `comparator` package:

| Operator | Constant | Description | Example |
|----------|----------|-------------|---------|
| `=` | `comparator.EQ` | Equal to | `Where("age", comparator.EQ, 25)` |
| `!=` | `comparator.NEQ` | Not equal to | `Where("status", comparator.NEQ, "inactive")` |
| `>` | `comparator.GT` | Greater than | `Where("age", comparator.GT, 18)` |
| `>=` | `comparator.GTE` | Greater than or equal | `Where("score", comparator.GTE, 80)` |
| `<` | `comparator.LT` | Less than | `Where("age", comparator.LT, 65)` |
| `<=` | `comparator.LTE` | Less than or equal | `Where("attempts", comparator.LTE, 3)` |
| `in` | `comparator.IN` | Value in array | `Where("role", comparator.IN, []any{"admin", "user"})` |
| `contains` | `comparator.CONTAINS` | String contains | `Where("email", comparator.CONTAINS, "@gmail.com")` |

## Performance Tips

1. **Use Count() for existence checks** instead of Find() when you only need to know if records exist
2. **Apply filters early** in the chain to reduce the dataset size
3. **Use Limit()** for large result sets to prevent memory issues
4. **Order results** consistently when using Offset() for pagination
5. **Consider using indices** on frequently queried fields in your Gremlin database

## Thread Safety

The query builder creates a new query instance for each operation and is safe to use concurrently. However, the underlying database connection should be managed appropriately for concurrent access.
