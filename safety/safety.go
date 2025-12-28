package safety

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/ikwerre-dev/emojidb/core"
	"github.com/ikwerre-dev/emojidb/crypto"
)

type SafetyBackup struct {
	Timestamp time.Time
	TableName string
	Data      core.Row
}

func BackupForSafety(db *core.Database, tableName string, row core.Row) error {
	return BatchBackupForSafety(db, tableName, []core.Row{row})
}

func BatchBackupForSafety(db *core.Database, tableName string, rows []core.Row) error {
	var buffer string
	for _, row := range rows {
		backup := SafetyBackup{
			Timestamp: time.Now(),
			TableName: tableName,
			Data:      row,
		}

		data, err := json.Marshal(backup)
		if err != nil {
			return err
		}

		encrypted, err := crypto.Encrypt(data, db.Key)
		if err != nil {
			return err
		}
		emojiPayload := crypto.EncodeToEmojis(encrypted)

		sizeBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(sizeBytes, uint32(len(encrypted)))
		sizeEncoded := crypto.EncodeToEmojis(sizeBytes)

		buffer += sizeEncoded + emojiPayload
	}

	db.Mu.Lock()
	defer db.Mu.Unlock()

	_, err := db.SafetyFile.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = db.SafetyFile.WriteString(buffer)
	if err != nil {
		return err
	}

	return db.SafetyFile.Sync()
}

func ListRecoveryPoints(db *core.Database) ([]time.Time, error) {
	db.Mu.RLock()
	defer db.Mu.RUnlock()

	var points []time.Time
	_, err := db.SafetyFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(db.SafetyFile)
	for {
		size, err := readIntEmoji(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// Read 'size' emojis
		payload := make([]byte, size)
		for i := 0; i < int(size); i++ {
			b, err := crypto.DecodeOne(br)
			if err != nil {
				return nil, err
			}
			payload[i] = b
		}

		decrypted, err := crypto.Decrypt(payload, db.Key)
		if err != nil {
			continue
		}

		var backup SafetyBackup
		if err := json.Unmarshal(decrypted, &backup); err != nil {
			continue
		}

		if time.Since(backup.Timestamp) <= 31*time.Minute {
			points = append(points, backup.Timestamp)
		}
	}

	return points, nil
}

func readIntEmoji(r *bufio.Reader) (uint32, error) {
	var buf []byte
	for i := 0; i < 4; i++ {
		b, err := crypto.DecodeOne(r)
		if err != nil {
			return 0, err
		}
		buf = append(buf, b)
	}
	return binary.LittleEndian.Uint32(buf), nil
}
