CREATE TABLE IF NOT EXISTS development_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title TEXT,
    status TEXT NOT NULL CHECK (status IN ('active', 'paused', 'ended')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    duration_seconds INTEGER CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
    source TEXT NOT NULL DEFAULT 'web' CHECK (source IN ('cli', 'web')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_development_sessions_user_started_at ON development_sessions(user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_development_sessions_project_started_at ON development_sessions(project_id, started_at DESC);
