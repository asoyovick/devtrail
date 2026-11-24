package cmd

import "fmt"

func PrintUsage() {
	fmt.Println(`DevTrail CLI

Usage:
  devtrail <command> [args]

Commands:
  init                         Initialize local DevTrail config and database
  login --email ... --password ... [--api ...]  Authenticate with DevTrail backend
  connect <directory>          Register and monitor a local project directory
  disconnect <project-id>      Disconnect a registered project (local only)
  projects                     List registered local projects
  status                       Show agent, project, and sync status
  watch                        Watch registered projects and record events
  sync                         Sync pending local events to DevTrail backend
  session start [--project id] [--title text]  Start a development session
  session pause                Pause active session
  session resume               Resume paused session
  session end                  End active session
  sessions                     List local sessions
  github connect|status        GitHub integration (backend-mediated; see status)

Environment:
  DEVTRAIL_HOME                Override config/data directory
`)
}
