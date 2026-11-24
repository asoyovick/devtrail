ALTER TABLE activity_events DROP CONSTRAINT IF EXISTS activity_events_source_check;
ALTER TABLE activity_events ADD CONSTRAINT activity_events_source_check CHECK (source IN ('cli', 'backend', 'web'));
