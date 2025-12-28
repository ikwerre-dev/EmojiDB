package tests

import (
	"os"
	"testing"

	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/query"
	"github.com/ikwerre-dev/emojidb/safety"
)

func TestQuery(t *testing.T) {
	dbPath := "test_query.db"
	safetyPath := dbPath + ".safety"

	// We don't defer remove here so the user can see the files
	os.Remove(dbPath)
	os.Remove(safetyPath)

	db, err := core.Open(dbPath, "secret")
	if err != nil {
		t.Fatalf("failed to open: %v", err)
	}
	defer db.Close()

	fields := []core.Field{
		{Name: "id", Type: core.FieldTypeInt},
		{Name: "name", Type: core.FieldTypeString},
		{Name: "age", Type: core.FieldTypeInt},
	}
	db.DefineSchema("users", fields)

	// Insert data
	db.Insert("users", core.Row{"id": 1, "name": "alice", "age": 30})
	db.Insert("users", core.Row{"id": 2, "name": "bob", "age": 25})

	if err := safety.BackupForSafety(db, "users", core.Row{"id": 1, "name": "alice", "age": 30}); err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if err := safety.BackupForSafety(db, "users", core.Row{"id": 2, "name": "bob", "age": 25}); err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Force flush to disk
	db.Flush("users")

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("test_query.db was not created")
	}
	if _, err := os.Stat(safetyPath); os.IsNotExist(err) {
		t.Error("safety.db was not created")
	}

	q := query.NewQuery(db, "users")
	results, err := q.Filter(func(r core.Row) bool {
		age, ok := r["age"].(int)
		return ok && age > 28
	}).Execute()

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1, got %d", len(results))
	}

	// Verify safety recovery points
	points, err := safety.ListRecoveryPoints(db)
	if err != nil {
		t.Fatalf("failed to list recovery points: %v", err)
	}
	if len(points) != 2 {
		t.Errorf("expected 2 recovery points, got %d", len(points))
	}
}
