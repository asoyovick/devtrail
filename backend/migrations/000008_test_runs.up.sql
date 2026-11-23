CREATE TABLE IF NOT EXISTS test_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    session_id UUID REFERENCES development_sessions(id) ON DELETE SET NULL,
    command TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('passed', 'failed', 'errored', 'skipped', 'unknown')),
    started_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    duration_ms BIGINT,
    total_tests INTEGER,
    passed_tests INTEGER,
    failed_tests INTEGER,
    failure_summary TEXT,
    event_id UUID UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_test_runs_user_started_at ON test_runs(user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_runs_project_started_at ON test_runs(project_id, started_at DESC);
