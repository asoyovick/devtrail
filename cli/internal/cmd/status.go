package cmd

import (
	"context"
	"fmt"

	"devtrail/cli/internal/config"
)

func Status(args []string) error {
	db, cfg, err := openDB()
	if err != nil { return err }
	defer db.Close()
	ctx := context.Background()
	projects, _ := db.ListProjects(ctx)
	var pending int
	_ = db.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE sync_status='pending'`).Scan(&pending)
	var synced int
	_ = db.SQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE sync_status='synced'`).Scan(&synced)
	dir, _ := config.Dir()
	fmt.Println("DevTrail status")
	fmt.Println("Config dir:", dir)
	fmt.Println("API base:  ", cfg.APIBaseURL)
	fmt.Println("Device ID: ", valueOr(cfg.DeviceID, "(not registered)"))
	fmt.Println("Projects:  ", len(projects))
	fmt.Println("Pending events:", pending)
	fmt.Println("Synced events: ", synced)
	for _, p := range projects {
		fmt.Printf("- %s (%s)\n", p.Name, p.ID)
	}
	return nil
}
