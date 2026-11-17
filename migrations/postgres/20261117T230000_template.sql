CREATE TABLE maintenance_template (
    id             TEXT NOT NULL PRIMARY KEY,
    name           TEXT NOT NULL UNIQUE,
    kind           TEXT NOT NULL CHECK (kind IN ('hours', 'days')),
    interval_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    interval_days  INTEGER NOT NULL DEFAULT 0,
    material       TEXT NOT NULL DEFAULT '',
    position       INTEGER NOT NULL DEFAULT 0
);

INSERT INTO maintenance_template (id, name, kind, interval_hours, interval_days, material, position) VALUES
    ('axes', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 1),
    ('connectors', 'Inspect the electrical connectors', 'hours', 600, 0, '', 2),
    ('rail', 'Lubricate the linear rail', 'days', 0, 90, '', 3),
    ('nozzle', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 4);
