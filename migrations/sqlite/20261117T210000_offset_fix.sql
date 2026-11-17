-- The first pass stored the printer's count as the distance, which counts the
-- hours this server had already seen twice. Mark the row so the module rebuilds
-- the distance from the count it holds and the hours it has tracked.
UPDATE maintenance_printer_hours SET offset_hours = -1;
