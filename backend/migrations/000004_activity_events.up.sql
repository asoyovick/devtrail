CREATE TABLE IF NOT EXISTS activity_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    session_id UUID REFERENCES development_sessions(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('cli', 'backend', 'web')),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    schema_version INTEGER NOT NULL DEFAULT 1,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activity_events_user_occurred_at ON activity_events(user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_events_project_occurred_at ON activity_events(project_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_events_session_occurred_at ON activity_events(session_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_events_type_occurred_at ON activity_events(event_type, occurred_at DESC);
