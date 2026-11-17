CREATE TABLE maintenance_seeded (
    device_id TEXT NOT NULL PRIMARY KEY,
    seeded_at TEXT NOT NULL
);

DELETE FROM maintenance_tasks
WHERE ctid NOT IN (
    SELECT MIN(ctid) FROM maintenance_tasks GROUP BY device_id, name
);
