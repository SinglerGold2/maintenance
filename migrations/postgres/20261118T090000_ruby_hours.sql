-- Diamond wears with printing, not with the calendar, so its interval counts
-- print hours like the other materials.
UPDATE maintenance_set_tasks SET interval_hours = 5000, interval_days = 0
WHERE material = 'ruby/diamond';
