-- Maintenance tasks: what an operator checks on a printer, how often, and when
-- it was last done. A time is an RFC 3339 text in UTC.

CREATE TABLE maintenance_tasks (
    id              TEXT NOT NULL PRIMARY KEY,
    device_id       TEXT NOT NULL,
    name            TEXT NOT NULL,
    kind            TEXT NOT NULL CHECK (kind IN ('hours', 'days')),
    interval_hours  DOUBLE PRECISION NOT NULL DEFAULT 0,
    interval_days   INTEGER NOT NULL DEFAULT 0,
    last_done_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_done_at    TEXT,
    created_at      TEXT NOT NULL
);

CREATE INDEX maintenance_tasks_device ON maintenance_tasks (device_id);
