-- The printer's counter is not a fixed number: it moves with every print, just
-- like the server's count. Store the distance between the two when the operator
-- types the printer's value in, and add it to the server's count from then on.

ALTER TABLE maintenance_printer_hours ADD COLUMN offset_hours REAL NOT NULL DEFAULT 0;

UPDATE maintenance_printer_hours SET offset_hours = total_hours;
