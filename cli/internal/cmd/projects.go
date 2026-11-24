package cmd

import (
	"context"
	"fmt"
)

func Projects(args []string) error {
	db, _, err := openDB()
	if err != nil { return err }
	defer db.Close()
	projects, err := db.ListProjects(context.Background())
	if err != nil { return err }
	if len(projects) == 0 {
		fmt.Println("No projects connected. Use: devtrail connect <directory>")
		return nil
	}
	for _, p := range projects {
		fmt.Printf("- %s\n  id: %s\n  path: %s\n  git: %s\n", p.Name, p.ID, p.Path, valueOr(p.RepoRoot, "(none)"))
	}
	return nil
}

func valueOr(v, fallback string) string { if v == "" { return fallback }; return v }
