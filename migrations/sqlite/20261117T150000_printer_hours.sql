-- The print hours the printer itself reports, typed in by the operator. The
-- server only sees the prints it recorded, so a machine that printed before
-- this server knew it carries a larger number on its own display.

CREATE TABLE maintenance_printer_hours (
    device_id   TEXT NOT NULL PRIMARY KEY,
    total_hours REAL NOT NULL,
    updated_at  TEXT NOT NULL
);
