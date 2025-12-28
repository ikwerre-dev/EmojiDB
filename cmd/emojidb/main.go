package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ikwerre-dev/EmojiDB/core"
	"github.com/ikwerre-dev/EmojiDB/query"
	"github.com/ikwerre-dev/EmojiDB/safety"
)

type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Response struct {
	ID    string      `json:"id"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

var db *core.Database

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(req.ID, "invalid json: "+err.Error())
			continue
		}

		handle(req)
	}
}

func handle(req Request) {
	switch req.Method {
	case "open":
		var p struct {
			Path string `json:"path"`
			Key  string `json:"key"`
		}
		json.Unmarshal(req.Params, &p)
		var err error
		db, err = core.Open(p.Path, p.Key)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "opened")
		}

	case "define_schema":
		var p struct {
			Table  string       `json:"table"`
			Fields []core.Field `json:"fields"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.DefineSchema(p.Table, p.Fields)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "defined")
		}

	case "sync_schema":
		var p struct {
			Table  string       `json:"table"`
			Fields []core.Field `json:"fields"`
			Force  bool         `json:"force"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.SyncSchema(p.Table, p.Fields, p.Force)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "migrated")
		}

	case "count":
		var p struct {
			Table string                 `json:"table"`
			Match map[string]interface{} `json:"match"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		count, err := db.Count(p.Table, p.Match)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, count)
		}

	case "drop_table":
		var p struct {
			Table string `json:"table"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.DropTable(p.Table)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "dropped")
		}

	case "pull_schema":
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.SaveSchemas()
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "pulled")
		}

	case "insert":
		var p struct {
			Table string   `json:"table"`
			Row   core.Row `json:"row"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.Insert(p.Table, p.Row)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "inserted")
		}

	case "update":
		var p struct {
			Table  string                 `json:"table"`
			Match  map[string]interface{} `json:"match"`
			Update core.Row               `json:"update"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := safety.Update(db, p.Table, func(r core.Row) bool {
			for k, v := range p.Match {
				if r[k] != v {
					return false
				}
			}
			return true
		}, p.Update)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "updated")
		}

	case "delete":
		var p struct {
			Table string                 `json:"table"`
			Match map[string]interface{} `json:"match"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := safety.Delete(db, p.Table, func(r core.Row) bool {
			for k, v := range p.Match {
				if r[k] != v {
					return false
				}
			}
			return true
		})
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "deleted")
		}

	case "batch_insert":
		var p struct {
			Table   string     `json:"table"`
			Records []core.Row `json:"records"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.BulkInsert(p.Table, p.Records)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "inserted")
		}

	case "query":
		var p struct {
			Table string `json:"table"`
			// Note: Complex filters via bridge are tricky.
			// For now, let's support a simple key-value filter.
			Match map[string]interface{} `json:"match"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}

		q := query.NewQuery(db, p.Table)
		if len(p.Match) > 0 {
			q = q.Filter(func(r core.Row) bool {
				for k, v := range p.Match {
					if r[k] != v {
						return false
					}
				}
				return true
			})
		}

		results, err := q.Execute()
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, results)
		}

	case "secure":
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.Secure()
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "secured")
		}

	case "rekey":
		var p struct {
			NewKey    string `json:"new_key"`
			MasterKey string `json:"master_key"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.ChangeKey(p.NewKey, p.MasterKey)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "rotated")
		}

	case "flush":
		var p struct {
			Table string `json:"table"`
		}
		json.Unmarshal(req.Params, &p)
		if db == nil {
			sendError(req.ID, "db not open")
			return
		}
		err := db.Flush(p.Table)
		if err != nil {
			sendError(req.ID, err.Error())
		} else {
			sendSuccess(req.ID, "flushed")
		}

	case "close":
		if db != nil {
			db.Close()
			sendSuccess(req.ID, "closed")
		} else {
			sendError(req.ID, "db not open")
		}

	default:
		sendError(req.ID, "unknown method")
	}
}

func sendSuccess(id string, data interface{}) {
	res, _ := json.Marshal(Response{ID: id, Data: data})
	fmt.Println(string(res))
}

func sendError(id string, err string) {
	res, _ := json.Marshal(Response{ID: id, Error: err})
	fmt.Println(string(res))
}
