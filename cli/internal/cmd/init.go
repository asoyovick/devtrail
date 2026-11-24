package cmd

import (
	"fmt"

	"devtrail/cli/internal/config"
)

func Init(args []string) error {
	cfg, err := config.Load()
	if err != nil { return err }
	if cfg.APIBaseURL == "" { cfg.APIBaseURL = "http://localhost:8080" }
	if err := config.Save(cfg); err != nil { return err }
	db, _, err := openDB()
	if err != nil { return err }
	defer db.Close()
	dir, _ := config.Dir()
	fmt.Println("DevTrail initialized")
	fmt.Println("Config dir:", dir)
	fmt.Println("API base:  ", cfg.APIBaseURL)
	return nil
}
