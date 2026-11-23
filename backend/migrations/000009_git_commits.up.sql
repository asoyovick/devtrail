CREATE TABLE IF NOT EXISTS git_commits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    session_id UUID REFERENCES development_sessions(id) ON DELETE SET NULL,
    commit_hash TEXT NOT NULL,
    branch TEXT,
    message_summary TEXT,
    author_name TEXT,
    author_email TEXT,
    committed_at TIMESTAMPTZ NOT NULL,
    files_changed INTEGER,
    insertions INTEGER,
    deletions INTEGER,
    event_id UUID UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, commit_hash)
);
CREATE INDEX IF NOT EXISTS idx_git_commits_user_committed_at ON git_commits(user_id, committed_at DESC);
CREATE INDEX IF NOT EXISTS idx_git_commits_project_committed_at ON git_commits(project_id, committed_at DESC);
