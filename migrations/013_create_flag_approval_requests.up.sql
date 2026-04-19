CREATE TABLE IF NOT EXISTS flag_approval_requests (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    flag_key      VARCHAR(255) NOT NULL,
    environment   VARCHAR(100) NOT NULL,
    change_type   VARCHAR(50)  NOT NULL,  -- toggle | update | delete
    payload       JSONB        NOT NULL DEFAULT '{}',
    requested_by  UUID         REFERENCES users(id),
    requested_by_name TEXT,
    status        VARCHAR(50)  NOT NULL DEFAULT 'pending',  -- pending | approved | rejected
    review_note   TEXT,
    reviewed_by   UUID         REFERENCES users(id),
    reviewed_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_approval_requests_project  ON flag_approval_requests(project_id);
CREATE INDEX IF NOT EXISTS idx_approval_requests_status   ON flag_approval_requests(status);
CREATE INDEX IF NOT EXISTS idx_approval_requests_flag_key ON flag_approval_requests(flag_key, environment);
