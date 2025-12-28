package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/query"
	"github.com/ikwerre-dev/emojidb/safety"
)

func TestFullShowcase(t *testing.T) {
	wd, _ := os.Getwd()
	dbPath := filepath.Join(wd, "showcase.db")
	dumpPath := filepath.Join(wd, "dump.json")
	safetyPath := filepath.Join(wd, "safety.db")

	key := "showcase-secret-2025"

	totalStart := time.Now()
	var timings []struct {
		name string
		took time.Duration
	}

	fmt.Println("\nSTARTING EMOJIDB FULL SHOWCASE")
	fmt.Println("==================================")

	// cleanup
	os.Remove(dbPath)
	os.Remove(safetyPath)
	os.Remove(dumpPath)

	// 1. Open Database
	start := time.Now()
	fmt.Printf("1. Opening Database (Encryption Mandatory)\n")
	db, err := core.Open(dbPath, key)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	defer db.Close()
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Open Database", time.Since(start)})

	// 2. Define Schema
	start = time.Now()
	fmt.Println("2. Defining Schema: 'products'")
	fields := []core.Field{
		{Name: "id", Type: core.FieldTypeInt},
		{Name: "name", Type: core.FieldTypeString},
		{Name: "price", Type: core.FieldTypeInt},
		{Name: "category", Type: core.FieldTypeString},
	}
	err = db.DefineSchema("products", fields)
	if err != nil {
		t.Fatalf("Failed to define schema: %v", err)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Define Schema", time.Since(start)})

	// 3. Ingestion (1500 records)
	start = time.Now()
	fmt.Println("3. Ingesting 1500 records into Hot Heap")
	for i := 1; i <= 1500; i++ {
		category := "tech"
		if i%3 == 0 {
			category = "food"
		} else if i%5 == 0 {
			category = "home"
		}

		row := core.Row{
			"id":       i,
			"name":     fmt.Sprintf("Product %d", i),
			"price":    i * 10,
			"category": category,
		}
		db.Insert("products", row)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Ingest 1500 Rows", time.Since(start)})

	// 4. Bulk Update (50 records)
	start = time.Now()
	fmt.Println("4. Safety Engine: Bulk Updating 50 records")
	err = safety.Update(db, "products", func(r core.Row) bool {
		id, ok := r["id"].(int)
		return ok && id >= 1100 && id < 1150
	}, core.Row{"category": "updated_bulk"})
	if err != nil {
		t.Fatalf("Bulk update failed: %v", err)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Bulk Update (50)", time.Since(start)})

	// 5. Single Updates (5 records)
	start = time.Now()
	fmt.Println("5. Safety Engine: Single Updating 5 records")
	for i := 1200; i < 1205; i++ {
		targetID := i
		safety.Update(db, "products", func(r core.Row) bool {
			id, _ := r["id"].(int)
			return id == targetID
		}, core.Row{"name": "Updated Single"})
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Single Update (5)", time.Since(start)})

	// 6. Bulk Delete (50 records)
	start = time.Now()
	fmt.Println("6. Safety Engine: Bulk Deleting 50 records")
	err = safety.Delete(db, "products", func(r core.Row) bool {
		id, ok := r["id"].(int)
		return ok && id >= 1300 && id < 1350
	})
	if err != nil {
		t.Fatalf("Bulk delete failed: %v", err)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Bulk Delete (50)", time.Since(start)})

	// 7. Single Deletes (5 records)
	start = time.Now()
	fmt.Println("7. Safety Engine: Single Deleting 5 records")
	for i := 1400; i < 1405; i++ {
		targetID := i
		safety.Delete(db, "products", func(r core.Row) bool {
			id, _ := r["id"].(int)
			return id == targetID
		})
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Single Delete (5)", time.Since(start)})

	// 8. Persistence (Flush)
	start = time.Now()
	fmt.Println("8. Flushing Hot Heap to Disk (Total Emoji Encoding)")
	db.Flush("products")
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Flush to Disk", time.Since(start)})

	// 9. Inspect File
	start = time.Now()
	fmt.Println("9. Inspecting Disk Content (Should be 100% Emojis)")
	content, _ := os.ReadFile(dbPath)
	if len(content) > 60 {
		fmt.Printf("   Preview: %s...\n", string(content[:60]))
	} else {
		fmt.Printf("   Content: %s\n", string(content))
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Inspect File", time.Since(start)})

	// 10. Query Engine
	start = time.Now()
	fmt.Println("10. Running Fluent Query: Category = 'updated_bulk'")
	results, err := query.NewQuery(db, "products").Filter(func(r core.Row) bool {
		cat, _ := r["category"].(string)
		return cat == "updated_bulk"
	}).Execute()

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("   Query Result: found %d matches\n", len(results))
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Execute Query", time.Since(start)})

	// 11. JSON Dump to File
	start = time.Now()
	fmt.Println("11. Dumping Table to dump.json")
	jsonDump, err := db.DumpAsJSON("products")
	if err != nil {
		t.Fatalf("Dump failed: %v", err)
	}

	err = os.WriteFile(dumpPath, []byte(jsonDump), 0644)
	if err != nil {
		t.Fatalf("Failed to write dump: %v", err)
	}
	fmt.Printf("   Saved to: %s\n", dumpPath)
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"JSON Export", time.Since(start)})

	fmt.Println("\nTEST SUMMARY")
	fmt.Println("==================================")
	for _, t := range timings {
		ms := float64(t.took.Nanoseconds()) / 1e6
		fmt.Printf("%-25s : %.3fms\n", t.name, ms)
	}
	fmt.Println("----------------------------------")
	totalMs := float64(time.Since(totalStart).Nanoseconds()) / 1e6
	fmt.Printf("%-25s : %.3fms\n", "TOTAL TIME", totalMs)
	fmt.Println("==================================")
}
