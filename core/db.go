package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ikwerre-dev/EmojiDB/crypto"
	"github.com/ikwerre-dev/EmojiDB/storage"
)

type Config struct {
	MemoryLimitMB   int
	ClumpSizeMB     int
	FlushIntervalMS int
}

type Database struct {
	Mu         sync.RWMutex
	Path       string
	Key        string
	File       *os.File
	SafetyFile *os.File
	SchemaFile *os.File
	Config     *Config
	Schemas    map[string]*Schema
	Tables     map[string]*Table
	Orphans    map[string][]*SealedClump
	SyncSafety bool
}

type Table struct {
	Mu            sync.RWMutex
	Db            *Database
	Name          string
	Schema        *Schema
	HotHeap       *HotHeap
	SealedClumps  []*SealedClump
	UniqueIndices map[string]map[interface{}]struct{}
}

func Open(path, key string) (*Database, error) {
	if key == "" {
		return nil, errors.New("database key is required")
	}

	// Ensure database resides in the 'emojidb' directory
	dir := "emojidb"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create emojidb directory: %v", err)
	}

	baseName := filepath.Base(path)
	fullPath := filepath.Join(dir, baseName)

	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	safetyPath := fullPath + ".safety"
	sFile, err := os.OpenFile(safetyPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		file.Close()
		return nil, err
	}

	schemaPath := fullPath + ".schema.json"
	schFile, err := os.OpenFile(schemaPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		file.Close()
		sFile.Close()
		return nil, err
	}

	db := &Database{
		Path:       fullPath,
		Key:        key,
		File:       file,
		SafetyFile: sFile,
		SchemaFile: schFile,
		Config:     &Config{},
		Schemas:    make(map[string]*Schema),
		Tables:     make(map[string]*Table),
		Orphans:    make(map[string][]*SealedClump),
		SyncSafety: true,
	}

	// Read header and load orphans/clumps
	if err := db.Load(); err != nil {
		file.Close()
		sFile.Close()
		schFile.Close()
		return nil, err
	}

	// Load persisted schemas
	if err := db.LoadSchemas(); err != nil {
		// Non-fatal if schema file is new/empty
	}

	return db, nil
}

func (db *Database) DefineSchema(tableName string, fields []Field) error {
	db.Mu.Lock()
	if db.Schemas == nil {
		db.Schemas = make(map[string]*Schema)
	}
	schema := &Schema{Version: 1, Fields: fields}
	db.Schemas[tableName] = schema

	indices := make(map[string]map[interface{}]struct{})
	for _, f := range fields {
		if f.Unique {
			indices[f.Name] = make(map[interface{}]struct{})
		}
	}

	if table, ok := db.Tables[tableName]; ok {
		table.Schema = schema
		table.UniqueIndices = indices
	} else {
		db.Tables[tableName] = &Table{
			Db:            db,
			Name:          tableName,
			Schema:        schema,
			HotHeap:       NewHotHeap(1000),
			SealedClumps:  make([]*SealedClump, 0),
			UniqueIndices: indices,
		}

		// Restore orphans if any
		if orphans, ok := db.Orphans[tableName]; ok {
			fmt.Printf("   Restoring %d clumps for table '%s'\n", len(orphans), tableName)
			db.Tables[tableName].SealedClumps = orphans
			// Populate unique indices from restored data
			for _, clump := range orphans {
				for _, row := range clump.Rows {
					for _, f := range fields {
						if f.Unique {
							val := row[f.Name]
							db.Tables[tableName].UniqueIndices[f.Name][val] = struct{}{}
						}
					}
				}
			}
			delete(db.Orphans, tableName)
		}
	}
	db.Mu.Unlock()

	return db.SaveSchemas()
}

func (db *Database) DiffSchema(tableName string, newFields []Field) ConflictReport {
	db.Mu.RLock()
	currentSchema, ok := db.Schemas[tableName]
	db.Mu.RUnlock()

	report := ConflictReport{Compatiable: true}

	if !ok {
		report.Conflicts = append(report.Conflicts, "TABLE_NEW: table does not exist on disk")
		return report
	}

	currentFieldMap := make(map[string]Field)
	for _, f := range currentSchema.Fields {
		currentFieldMap[f.Name] = f
	}

	newFieldMap := make(map[string]Field)
	for _, f := range newFields {
		newFieldMap[f.Name] = f
		if oldF, exists := currentFieldMap[f.Name]; exists {
			if oldF.Type != f.Type {
				report.Compatiable = false
				report.Conflicts = append(report.Conflicts, fmt.Sprintf("TYPE_MISMATCH: field '%s' change from %v to %v", f.Name, oldF.Type, f.Type))
			}
		} else {
			report.Conflicts = append(report.Conflicts, fmt.Sprintf("FIELD_ADD: new field '%s' will be added", f.Name))
		}
	}

	for oldName := range currentFieldMap {
		if _, exists := newFieldMap[oldName]; !exists {
			report.Destructive = true
			report.Conflicts = append(report.Conflicts, fmt.Sprintf("FIELD_REMOVE: field '%s' and its data will be inaccessible", oldName))
		}
	}

	return report
}

func (db *Database) SyncSchema(tableName string, newFields []Field) error {
	report := db.DiffSchema(tableName, newFields)
	if !report.Compatiable {
		return fmt.Errorf("incompatible schema change: %v", report.Conflicts)
	}

	db.Mu.Lock()
	schema := &Schema{Version: 1, Fields: newFields}
	db.Schemas[tableName] = schema

	if table, ok := db.Tables[tableName]; ok {
		table.Schema = schema
		// Update unique indices
		indices := make(map[string]map[interface{}]struct{})
		for _, f := range newFields {
			if f.Unique {
				indices[f.Name] = make(map[interface{}]struct{})
			}
		}
		table.UniqueIndices = indices

		// Recalculate unique indices from existing data
		for _, clump := range table.SealedClumps {
			for _, row := range clump.Rows {
				for _, f := range newFields {
					if f.Unique {
						val, ok := row[f.Name]
						if ok {
							table.UniqueIndices[f.Name][val] = struct{}{}
						}
					}
				}
			}
		}
		for _, row := range table.HotHeap.Rows {
			for _, f := range newFields {
				if f.Unique {
					val, ok := row[f.Name]
					if ok {
						table.UniqueIndices[f.Name][val] = struct{}{}
					}
				}
			}
		}
	}
	db.Mu.Unlock()

	return db.SaveSchemas()
}

func (db *Database) Insert(tableName string, record Row) error {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return errors.New("table not found: " + tableName)
	}

	table.Mu.Lock()
	defer table.Mu.Unlock()

	// Check constraints
	for _, field := range table.Schema.Fields {
		val, ok := record[field.Name]
		if !ok {
			return errors.New("missing field: " + field.Name)
		}

		if field.Unique {
			if _, exists := table.UniqueIndices[field.Name][val]; exists {
				return errors.New("unique constraint violation: " + field.Name)
			}
		}
	}

	// Apply unique indices
	for _, field := range table.Schema.Fields {
		if field.Unique {
			table.UniqueIndices[field.Name][record[field.Name]] = struct{}{}
		}
	}

	table.HotHeap.Rows = append(table.HotHeap.Rows, record)

	if len(table.HotHeap.Rows) >= table.HotHeap.MaxRows {
		// Auto-flush
		// Actually table.Mu is held.
		// Let's do a safe auto-flush.
		clump := &SealedClump{
			Rows:     table.HotHeap.Rows,
			SealedAt: time.Now(),
			Metadata: ClumpMetadata{
				RowCount:      len(table.HotHeap.Rows),
				CreatedAt:     table.HotHeap.CreatedAt,
				SchemaVersion: table.Schema.Version,
			},
		}
		table.SealedClumps = append(table.SealedClumps, clump)
		table.HotHeap = NewHotHeap(1000)

		// Persistence happens outside table lock to avoid deadlocks with db.Mu
		go db.PersistClump(tableName, clump)
	}

	return nil
}

func (db *Database) BulkInsert(tableName string, records []Row) error {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return errors.New("table not found: " + tableName)
	}

	table.Mu.Lock()
	defer table.Mu.Unlock()

	// 1. Validation Phase (All or Nothing)
	for i, record := range records {
		for _, field := range table.Schema.Fields {
			val, ok := record[field.Name]
			if !ok {
				return fmt.Errorf("row %d: missing field: %s", i, field.Name)
			}
			if field.Unique {
				if _, exists := table.UniqueIndices[field.Name][val]; exists {
					return fmt.Errorf("row %d: unique constraint violation: %s", i, field.Name)
				}
				// Also check against other rows in this batch to prevent duplicates within the batch
				for j := 0; j < i; j++ {
					if records[j][field.Name] == val {
						return fmt.Errorf("row %d: duplicate value in batch for field: %s", i, field.Name)
					}
				}
			}
		}
	}

	// 2. Application Phase
	for _, record := range records {
		for _, field := range table.Schema.Fields {
			if field.Unique {
				table.UniqueIndices[field.Name][record[field.Name]] = struct{}{}
			}
		}
		table.HotHeap.Rows = append(table.HotHeap.Rows, record)
	}

	// Check for auto-flush once at the end
	if len(table.HotHeap.Rows) >= table.HotHeap.MaxRows {
		clump := &SealedClump{
			Rows:     table.HotHeap.Rows,
			SealedAt: time.Now(),
			Metadata: ClumpMetadata{
				RowCount:      len(table.HotHeap.Rows),
				CreatedAt:     table.HotHeap.CreatedAt,
				SchemaVersion: table.Schema.Version,
			},
		}
		table.SealedClumps = append(table.SealedClumps, clump)
		table.HotHeap = NewHotHeap(1000)
		go db.PersistClump(tableName, clump)
	}

	return nil
}

func (db *Database) PersistClump(tableName string, clump *SealedClump) error {
	return storage.PersistClump(db.File, &db.Mu, tableName, clump, db.Key, crypto.Encrypt, crypto.EncodeToEmojis)
}

func (db *Database) Load() error {
	handleClump := func(tableName string, data []byte) error {
		var clump SealedClump
		if err := json.Unmarshal(data, &clump); err != nil {
			return err
		}
		db.Mu.Lock()
		table, ok := db.Tables[tableName]
		if ok {
			table.SealedClumps = append(table.SealedClumps, &clump)
		} else {
			db.Orphans[tableName] = append(db.Orphans[tableName], &clump)
		}
		db.Mu.Unlock()
		return nil
	}

	return storage.Load(db.File, &db.Mu, db.Key, crypto.Decrypt, handleClump)
}

func (db *Database) Secure() error {
	path := filepath.Join(filepath.Dir(db.Path), "secure.pem")
	if _, err := os.Stat(path); err == nil {
		return errors.New("security already initialized")
	}

	rawKey := make([]byte, 32)
	crypto.RandRead(rawKey)
	emojiKey := crypto.EncodeToEmojis(rawKey)

	return os.WriteFile(path, []byte(emojiKey), 0600)
}

func (db *Database) ChangeKey(newKey string, masterKey string) error {
	path := filepath.Join(filepath.Dir(db.Path), "secure.pem")
	actualMaster, err := os.ReadFile(path)
	if err != nil {
		return errors.New("security not initialized or secure.pem missing")
	}

	if string(actualMaster) != masterKey {
		return errors.New("invalid master key provided")
	}

	db.Mu.Lock()
	defer db.Mu.Unlock()

	// 1. Truncate and reset file
	err = db.File.Truncate(0)
	if err != nil {
		return err
	}
	_, err = db.File.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// 2. Write header
	if err := storage.WriteHeader(db.File); err != nil {
		return err
	}

	oldKey := db.Key
	db.Key = newKey

	// 3. Re-persist all existing sealed clumps with new key
	for tableName, table := range db.Tables {
		table.Mu.RLock()
		for _, clump := range table.SealedClumps {
			if err := storage.InternalPersistClump(db.File, tableName, clump, db.Key, crypto.Encrypt, crypto.EncodeToEmojis); err != nil {
				db.Key = oldKey // Rollback
				table.Mu.RUnlock()
				return err
			}
		}
		table.Mu.RUnlock()
	}

	return db.File.Sync()
}

func (db *Database) Flush(tableName string) error {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return errors.New("table not found")
	}

	table.Mu.Lock()
	if len(table.HotHeap.Rows) == 0 {
		table.Mu.Unlock()
		return nil
	}

	clump := &SealedClump{
		Rows:     table.HotHeap.Rows,
		SealedAt: time.Now(),
		Metadata: ClumpMetadata{
			RowCount:      len(table.HotHeap.Rows),
			CreatedAt:     table.HotHeap.CreatedAt,
			SchemaVersion: table.Schema.Version,
		},
	}
	table.SealedClumps = append(table.SealedClumps, clump)
	table.HotHeap = NewHotHeap(1000)
	table.Mu.Unlock()

	return storage.PersistClump(db.File, &db.Mu, tableName, clump, db.Key, crypto.Encrypt, crypto.EncodeToEmojis)
}

func (db *Database) DumpAsJSON(tableName string) (string, error) {
	db.Mu.RLock()
	table, ok := db.Tables[tableName]
	db.Mu.RUnlock()

	if !ok {
		return "", errors.New("table not found")
	}

	table.Mu.RLock()
	defer table.Mu.RUnlock()

	var allRows []Row
	for _, clump := range table.SealedClumps {
		allRows = append(allRows, clump.Rows...)
	}
	if table.HotHeap != nil {
		allRows = append(allRows, table.HotHeap.Rows...)
	}

	data, err := json.MarshalIndent(allRows, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (db *Database) SaveSchemas() error {
	db.Mu.RLock()
	data, err := json.MarshalIndent(db.Schemas, "", "  ")
	db.Mu.RUnlock()
	if err != nil {
		return err
	}

	db.Mu.Lock()
	defer db.Mu.Unlock()

	if err := db.SchemaFile.Truncate(0); err != nil {
		return err
	}
	if _, err := db.SchemaFile.Seek(0, 0); err != nil {
		return err
	}
	_, err = db.SchemaFile.Write(data)
	return db.SchemaFile.Sync()
}

func (db *Database) LoadSchemas() error {
	db.Mu.Lock()
	defer db.Mu.Unlock()

	if _, err := db.SchemaFile.Seek(0, 0); err != nil {
		return err
	}

	content, err := os.ReadFile(db.Path + ".schema.json")
	if err != nil || len(content) == 0 {
		return nil
	}

	var schemas map[string]*Schema
	if err := json.Unmarshal(content, &schemas); err != nil {
		return err
	}

	db.Schemas = schemas
	// Also re-initialize tables from schemas
	for name, schema := range schemas {
		if _, ok := db.Tables[name]; !ok {
			// We skip calling db.DefineSchema recursively and just init the table maps
			indices := make(map[string]map[interface{}]struct{})
			for _, f := range schema.Fields {
				if f.Unique {
					indices[f.Name] = make(map[interface{}]struct{})
				}
			}
			db.Tables[name] = &Table{
				Db:            db,
				Name:          name,
				Schema:        schema,
				HotHeap:       NewHotHeap(1000),
				SealedClumps:  make([]*SealedClump, 0),
				UniqueIndices: indices,
			}
			// Restore orphans if any
			if orphans, ok := db.Orphans[name]; ok {
				db.Tables[name].SealedClumps = orphans
				for _, clump := range orphans {
					for _, row := range clump.Rows {
						for _, f := range schema.Fields {
							if f.Unique {
								val := row[f.Name]
								db.Tables[name].UniqueIndices[f.Name][val] = struct{}{}
							}
						}
					}
				}
				delete(db.Orphans, name)
			}
		}
	}

	return nil
}

func (db *Database) Close() error {
	db.Mu.Lock()
	defer db.Mu.Unlock()
	if db.SafetyFile != nil {
		db.SafetyFile.Close()
	}
	if db.SchemaFile != nil {
		db.SchemaFile.Close()
	}
	if db.File != nil {
		return db.File.Close()
	}
	return nil
}
