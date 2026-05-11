-- Add name and site_url to projects
ALTER TABLE projects ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN IF NOT EXISTS site_url TEXT NOT NULL DEFAULT '';

-- Add status column: 1=active, 2=disabled (by partner), 3=blocked (by Duel Duck)
ALTER TABLE projects ADD COLUMN IF NOT EXISTS status SMALLINT NOT NULL DEFAULT 1;

-- Backfill status from existing is_blocked flag
UPDATE projects SET status = 3 WHERE is_blocked = true;
UPDATE projects SET status = 1 WHERE is_blocked = false;

-- project_status_history: audit trail of project status transitions
CREATE TABLE IF NOT EXISTS project_status_history (
    id         BIGSERIAL PRIMARY KEY,
    project_id UUID      NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    status     SMALLINT  NOT NULL,
    changed_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_project_status_history_project_id
    ON project_status_history (project_id, changed_at);

-- Backfill history: one initial entry per project
INSERT INTO project_status_history (project_id, status, changed_at)
SELECT id, status, created_at FROM projects;
