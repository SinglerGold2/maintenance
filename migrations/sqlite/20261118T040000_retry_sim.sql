-- The simulated printer's prints were marked read while the module could not
-- see them as the person reading the page. Forget those marks and read them
-- again.
DELETE FROM maintenance_jobs_seen WHERE device_id LIKE '7d0932c6%';
