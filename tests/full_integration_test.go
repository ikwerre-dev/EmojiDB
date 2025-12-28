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

	// 4. Safety Engine (Update & Backup)
	start = time.Now()
	fmt.Println("4. Safety Engine: Updating Product 1 price (Automatic Backup)")
	err = safety.Update(db, "products", func(r core.Row) bool {
		id, ok := r["id"].(int)
		return ok && id == 1
	}, core.Row{"price": 99})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Safety Update", time.Since(start)})

	// 5. Persistence (Flush)
	start = time.Now()
	fmt.Println("5. Flushing Hot Heap to Disk (Total Emoji Encoding)")
	db.Flush("products")
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Flush to Disk", time.Since(start)})

	// 6. Inspect File
	start = time.Now()
	fmt.Println("6. Inspecting Disk Content (Should be 100% Emojis)")
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

	// 7. Query Engine
	start = time.Now()
	fmt.Println("7. Running Fluent Query: Category = 'tech' && Price < 100")
	results, err := query.NewQuery(db, "products").Filter(func(r core.Row) bool {
		cat, _ := r["category"].(string)
		price, _ := r["price"].(int)
		return cat == "tech" && price < 100
	}).Execute()

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	fmt.Printf("   Query Result: found %d matches\n", len(results))
	timings = append(timings, struct {
		name string
		took time.Duration
	}{"Execute Query", time.Since(start)})

	// 8. JSON Dump to File
	start = time.Now()
	fmt.Println("8. Dumping Table to dump.json")
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

	fmt.Println("\nSHOWCASE SUMMARY")
	fmt.Println("==================================")
	for _, t := range timings {
		fmt.Printf("%-25s : %dms\n", t.name, t.took.Milliseconds())
	}
	fmt.Println("----------------------------------")
	fmt.Printf("%-25s : %dms\n", "TOTAL TIME", time.Since(totalStart).Milliseconds())
	fmt.Println("==================================")
}
