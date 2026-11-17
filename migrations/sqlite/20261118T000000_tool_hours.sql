-- The print history has no per-tool time, so the hours each tool ran are read
-- from the G-code of every ended print once: the file says how much filament
-- each tool pushed, and the print's length is split by that share.

CREATE TABLE maintenance_jobs_seen (
    device_id TEXT NOT NULL,
    job_id    TEXT NOT NULL,
    seen_at   TEXT NOT NULL,
    PRIMARY KEY (device_id, job_id)
);

CREATE TABLE maintenance_tool_hours (
    device_id  TEXT NOT NULL,
    tool       INTEGER NOT NULL,
    hours      REAL NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL,
    PRIMARY KEY (device_id, tool)
);
