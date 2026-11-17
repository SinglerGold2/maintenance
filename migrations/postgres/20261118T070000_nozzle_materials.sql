-- The nozzle check is not one number: the material decides it. Brass wears
-- fastest, diamond lasts a year.
UPDATE maintenance_set_tasks
SET kind = 'nozzle', material = 'hardened steel', interval_hours = 1500, interval_days = 0
WHERE name = 'Check the nozzle, replace when worn';
