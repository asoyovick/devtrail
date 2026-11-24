package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"devtrail/cli/internal/localdb"
	"github.com/fsnotify/fsnotify"
)

var ignoreDirs = map[string]bool{".git": true, "node_modules": true, "vendor": true, ".next": true, "dist": true, "build": true, ".cache": true, "__pycache__": true, ".idea": true, ".vscode": true}
var ignoreFiles = map[string]bool{".env": true, ".env.local": true, ".DS_Store": true}

func Watch(args []string) error {
	db, cfg, err := openDB()
	if err != nil { return err }
	defer db.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	projects, err := db.ListProjects(ctx)
	if err != nil { return err }
	if len(projects) == 0 { return fmt.Errorf("no projects connected; run: devtrail connect <directory>") }

	watcher, err := fsnotify.NewWatcher()
	if err != nil { return err }
	defer watcher.Close()

	for _, p := range projects {
		if err := addRecursive(watcher, p.Path); err != nil {
			fmt.Fprintf(os.Stderr, "warn: watch %s: %v\n", p.Path, err)
		} else {
			fmt.Println("Watching:", p.Path)
		}
	}
	fmt.Println("Recording events locally. Press Ctrl+C to stop.")

	// debounce map
	type pending struct { ev fsnotify.Event; at time.Time }
	pendingMap := map[string]pending{}
	ticker := time.NewTicker(750 * time.Millisecond)
	syncTicker := time.NewTicker(10 * time.Second)
	defer ticker.Stop(); defer syncTicker.Stop()

	flush := func() {
		now := time.Now()
		for k, p := range pendingMap {
			if now.Sub(p.at) < 500*time.Millisecond { continue }
			delete(pendingMap, k)
			handleFSEvent(ctx, db, projects, p.ev)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			fmt.Println("\nStopping watch.")
			return nil
		case err, ok := <-watcher.Errors:
			if !ok { return nil }
			fmt.Fprintln(os.Stderr, "watch error:", err)
		case ev, ok := <-watcher.Events:
			if !ok { return nil }
			if ev.Name == "" { continue }
			if shouldIgnore(ev.Name) { continue }
			// if new dir created, add watch
			if ev.Has(fsnotify.Create) {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() && !shouldIgnore(ev.Name) {
					_ = addRecursive(watcher, ev.Name)
				}
			}
			pendingMap[ev.Op.String()+"|"+ev.Name] = pending{ev: ev, at: time.Now()}
		case <-ticker.C:
			flush()
		case <-syncTicker.C:
			if cfg.AuthToken != "" || cfg.APIBaseURL != "" {
				if err := syncPending(ctx, db, cfg); err != nil {
					fmt.Fprintln(os.Stderr, "sync:", err)
				}
			}
		}
	}
}

func addRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil { return nil }
		if d.IsDir() {
			if shouldIgnore(path) { return filepath.SkipDir }
			if err := w.Add(path); err != nil { return nil }
		}
		return nil
	})
}

func shouldIgnore(path string) bool {
	base := filepath.Base(path)
	if ignoreDirs[base] || ignoreFiles[base] { return true }
	if strings.HasPrefix(base, ".env") { return true }
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, p := range parts { if ignoreDirs[p] { return true } }
	return false
}

func handleFSEvent(ctx context.Context, db *localdb.DB, projects []localdb.Project, ev fsnotify.Event) {
	proj, rel, ok := matchProject(projects, ev.Name)
	if !ok { return }
	eventType := "file.modified"
	switch {
	case ev.Has(fsnotify.Create): eventType = "file.created"
	case ev.Has(fsnotify.Remove): eventType = "file.deleted"
	case ev.Has(fsnotify.Rename): eventType = "file.renamed"
	case ev.Has(fsnotify.Write): eventType = "file.modified"
	case ev.Has(fsnotify.Chmod): return
	}
	payload, _ := json.Marshal(map[string]any{"path": rel, "op": ev.Op.String(), "project_name": proj.Name})
	e := localdb.Event{EventID: uuidv4(), ProjectID: proj.ID, EventType: eventType, OccurredAt: time.Now().UTC(), Source: "local_agent", Payload: payload, SchemaVersion: 1}
	if err := db.InsertEvent(ctx, e); err != nil { fmt.Fprintln(os.Stderr, "insert event:", err); return }
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), eventType, rel)
}

func matchProject(projects []localdb.Project, path string) (localdb.Project, string, bool) {
	var best localdb.Project
	bestLen := -1
	for _, p := range projects {
		if path == p.Path || strings.HasPrefix(path, p.Path+string(os.PathSeparator)) {
			if len(p.Path) > bestLen { best = p; bestLen = len(p.Path) }
		}
	}
	if bestLen < 0 { return localdb.Project{}, "", false }
	rel, err := filepath.Rel(best.Path, path)
	if err != nil { rel = path }
	return best, filepath.ToSlash(rel), true
}
