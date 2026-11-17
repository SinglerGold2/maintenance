ALTER TABLE maintenance_printer_hours ADD COLUMN offset_hours DOUBLE PRECISION NOT NULL DEFAULT 0;

UPDATE maintenance_printer_hours SET offset_hours = total_hours;
