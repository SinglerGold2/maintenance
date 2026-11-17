CREATE TABLE maintenance_printer_hours (
    device_id   TEXT NOT NULL PRIMARY KEY,
    total_hours DOUBLE PRECISION NOT NULL,
    updated_at  TEXT NOT NULL
);
