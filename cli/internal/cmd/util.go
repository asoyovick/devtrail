package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"devtrail/cli/internal/config"
	"devtrail/cli/internal/localdb"
)

func openDB() (*localdb.DB, config.Config, error) {
	cfg, err := config.Load()
	if err != nil { return nil, config.Config{}, err }
	dbPath, err := config.DBPath()
	if err != nil { return nil, config.Config{}, err }
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil { return nil, config.Config{}, err }
	db, err := localdb.Open(dbPath)
	if err != nil { return nil, config.Config{}, err }
	return db, cfg, nil
}

func uuidv4() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil { panic(err) }
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
