CREATE TABLE IF NOT EXISTS decisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    context TEXT NOT NULL,
    problem TEXT NOT NULL,
    alternatives TEXT,
    selected_approach TEXT NOT NULL,
    reasoning TEXT NOT NULL,
    consequences TEXT,
    status TEXT NOT NULL DEFAULT 'accepted' CHECK (status IN ('proposed', 'accepted', 'superseded', 'deprecated')),
    decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_decisions_user_created_at ON decisions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_decisions_project_created_at ON decisions(project_id, created_at DESC);
