-- The seeded schedule is written once per printer. Earlier releases wrote it
-- whenever a load found no task, and two loads at the same time wrote it twice.
-- This table is the guard, and the duplicate rows already written go away here.

CREATE TABLE maintenance_seeded (
    device_id TEXT NOT NULL PRIMARY KEY,
    seeded_at TEXT NOT NULL
);

DELETE FROM maintenance_tasks
WHERE rowid NOT IN (
    SELECT MIN(rowid) FROM maintenance_tasks GROUP BY device_id, name
);
