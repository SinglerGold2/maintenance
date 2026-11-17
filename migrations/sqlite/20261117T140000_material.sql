-- A task may name the nozzle material it tracks, so a nozzle task carries the
-- wear class the operator picked. Empty for every task that is not about a
-- nozzle.

ALTER TABLE maintenance_tasks ADD COLUMN material TEXT NOT NULL DEFAULT '';
