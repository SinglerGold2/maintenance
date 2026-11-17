-- A task the module ships is not the operator's to remove. This column says
-- which ones those are; the rows already written by the seed are marked here.

ALTER TABLE maintenance_tasks ADD COLUMN builtin INTEGER NOT NULL DEFAULT 0;

UPDATE maintenance_tasks SET builtin = 1 WHERE name IN (
    'Check axes, belts, fans and extruder',
    'Inspect the electrical connectors',
    'Lubricate the linear rail',
    'Check the nozzle, replace when worn',
    'Inspect the machine'
);
