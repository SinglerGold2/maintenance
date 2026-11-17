-- 0.10.x marked a print seen even when it could not read it, so the hours of
-- the prints it failed on were lost. Forget the marks and read them again.
DELETE FROM maintenance_jobs_seen;
