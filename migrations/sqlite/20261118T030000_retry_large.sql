-- 0.10.11 skipped a print whose file was larger than its limit and marked it
-- read, so its tool hours were lost. Forget that one mark; the print is read
-- again with the higher limit.
DELETE FROM maintenance_jobs_seen WHERE job_id = '01a1ea56-3653-7be9-a938-b58298a38c46';
