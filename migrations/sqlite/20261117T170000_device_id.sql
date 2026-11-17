-- The seed wrote its tasks without the printer's id, so they belonged to no
-- printer and every load seeded again. Remove those orphans; new tasks carry
-- the id.
DELETE FROM maintenance_tasks WHERE device_id = '';
