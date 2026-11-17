-- A task's counter is held in the printer's own hours, not in what this server
-- has seen. The rows written so far were seeded, so they start at the printer's
-- zero and their elapsed time is the whole count the machine shows.
UPDATE maintenance_tasks SET last_done_hours = 0 WHERE builtin = 1;
