# EmojiDB

EmojiDB is a memory-first, encrypted, emoji-encoded embedded database engine written in Go.

## Features
- **Memory-First**: High-speed ingestion using Hot Heaps.
- **Encrypted**: AES-GCM encryption for all stored data.
- **Emoji Encoding**: Encrypted bytes are mapped to a deterministic set of 256 emojis.
- **Safety Engine**: Automatic 30-minute recovery window for updates and deletes.
- **Fluent Query API**: Simple, chainable filtering and projection.

## Usage

```go
package main

import (
	"fmt"
	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/query"
)

func main() {
	// Open database
	db, _ := core.Open("data.db", "your-secret-key", true)
	defer db.Close()

	// Define schema
	fields := []core.Field{
		{Name: "id", Type: core.FieldTypeInt},
		{Name: "name", Type: core.FieldTypeString},
	}
	db.DefineSchema("users", fields)

	// Insert data
	db.Insert("users", core.Row{"id": 1, "name": "alice"})

	// Query data
	q := query.NewQuery(db, "users")
	results, _ := q.Filter(func(r core.Row) bool {
		return r["name"] == "alice"
	}).Execute()

	fmt.Println(results)
}
```

## Testing

Run all tests across all modules:

```bash
go test -v ./...
```

To run specific tests:

```bash
# Core tests
go test -v ./core/...

# Safety engine tests
go test -v ./safety/...

# Query engine tests
go test -v ./query/...

# Integration tests
go test -v ./tests/...
```

## Structure
- `core/`: Core engine and memory logic.
- `crypto/`: Encryption and emoji mapping.
- `storage/`: Persistence and file handling.
- `query/`: Fluent query engine.
- `safety/`: Backup and recovery engine.
- `tests/`: Integration and cross-module tests.
