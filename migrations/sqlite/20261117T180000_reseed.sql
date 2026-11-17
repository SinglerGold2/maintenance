-- The marker outlived the rows it guarded: the orphans were removed and the
-- marker stayed, so the schedule was never written again. Clear it once, and
-- the next load seeds the printer properly.
DELETE FROM maintenance_seeded;
