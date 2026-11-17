CREATE TABLE maintenance_jobs_seen (
    device_id TEXT NOT NULL,
    job_id    TEXT NOT NULL,
    seen_at   TEXT NOT NULL,
    PRIMARY KEY (device_id, job_id)
);

CREATE TABLE maintenance_tool_hours (
    device_id  TEXT NOT NULL,
    tool       INTEGER NOT NULL,
    hours      DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (device_id, tool)
);
