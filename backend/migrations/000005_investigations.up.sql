CREATE TABLE IF NOT EXISTS investigations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    session_id UUID REFERENCES development_sessions(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    problem_description TEXT NOT NULL,
    observed_behavior TEXT,
    investigation_steps TEXT,
    attempted_approaches TEXT,
    root_cause TEXT,
    resolution TEXT,
    lessons_learned TEXT,
    prevention_strategies TEXT,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'abandoned')),
    severity TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_investigations_user_created_at ON investigations(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_investigations_project_created_at ON investigations(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS investigation_related_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investigation_id UUID NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS investigation_links (
    investigation_id UUID NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    related_investigation_id UUID NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL DEFAULT 'related' CHECK (relation_type IN ('related', 'duplicate', 'regression')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (investigation_id, related_investigation_id),
    CHECK (investigation_id <> related_investigation_id)
);
