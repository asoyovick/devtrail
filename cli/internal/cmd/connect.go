package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devtrail/cli/internal/localdb"
)

func Connect(args []string) error {
	if len(args) < 1 { return fmt.Errorf("usage: devtrail connect <directory>") }
	dir := args[0]
	abs, err := filepath.Abs(dir)
	if err != nil { return err }
	info, err := os.Stat(abs)
	if err != nil { return err }
	if !info.IsDir() { return fmt.Errorf("not a directory: %s", abs) }
	repoRoot := findGitRoot(abs)
	db, _, err := openDB()
	if err != nil { return err }
	defer db.Close()
	ctx := context.Background()
	if existing, err := db.GetProjectByPath(ctx, abs); err == nil {
		fmt.Println("Project already connected:", existing.Name)
		fmt.Println("ID:", existing.ID)
		fmt.Println("Path:", existing.Path)
		return nil
	}
	name := filepath.Base(abs)
	p := localdb.Project{ID: uuidv4(), Name: name, Path: abs, RepoRoot: repoRoot, CreatedAt: time.Now().UTC()}
	if err := db.UpsertProject(ctx, p); err != nil { return err }
	fmt.Println("Connected project")
	fmt.Println("ID:      ", p.ID)
	fmt.Println("Name:    ", p.Name)
	fmt.Println("Path:    ", p.Path)
	if repoRoot != "" { fmt.Println("Git root:", repoRoot) } else { fmt.Println("Git root: (not a git repository)") }
	fmt.Println("Monitoring will record events when you run: devtrail watch")
	return nil
}

func Disconnect(args []string) error {
	if len(args) < 1 { return fmt.Errorf("usage: devtrail disconnect <project-id-or-path>") }
	key := args[0]
	db, _, err := openDB()
	if err != nil { return err }
	defer db.Close()
	ctx := context.Background()
	res, err := db.SQL.ExecContext(ctx, `DELETE FROM projects WHERE id=? OR path=?`, key, key)
	if err != nil { return err }
	n, _ := res.RowsAffected()
	if n == 0 { return fmt.Errorf("project not found: %s", key) }
	fmt.Println("Disconnected project:", key)
	return nil
}

func findGitRoot(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil { return dir }
		parent := filepath.Dir(dir)
		if parent == dir { return "" }
		dir = parent
	}
}
