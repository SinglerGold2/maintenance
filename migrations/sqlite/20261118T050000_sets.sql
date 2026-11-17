-- Maintenance sets: a named schedule a printer can be assigned to. The
-- shipped sets start from the Prusa Knowledge Base and can be edited, switched
-- off or duplicated; a printer takes its tasks from the set it points at.

CREATE TABLE IF NOT EXISTS maintenance_sets (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL,
    model    TEXT NOT NULL DEFAULT '*',
    builtin  INTEGER NOT NULL DEFAULT 0,
    enabled  INTEGER NOT NULL DEFAULT 1,
    position INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS maintenance_set_tasks (
    id             TEXT PRIMARY KEY,
    set_id         TEXT NOT NULL REFERENCES maintenance_sets (id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    kind           TEXT NOT NULL DEFAULT 'hours',
    interval_hours REAL NOT NULL DEFAULT 0,
    interval_days  INTEGER NOT NULL DEFAULT 0,
    material       TEXT NOT NULL DEFAULT '',
    position       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS maintenance_printer_sets (
    device_id TEXT PRIMARY KEY,
    set_id    TEXT NOT NULL
);

INSERT INTO maintenance_sets (id, name, model, builtin, enabled, position) VALUES ('prusa-kb-coreone', 'Prusa KB - Core One', 'COREONE', 1, 1, 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreone-t1', 'prusa-kb-coreone', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreone-t2', 'prusa-kb-coreone', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreone-t3', 'prusa-kb-coreone', 'Lubricate the linear rail', 'days', 0, 90, '', 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreone-t4', 'prusa-kb-coreone', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 3) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_sets (id, name, model, builtin, enabled, position) VALUES ('prusa-kb-coreonel', 'Prusa KB - Core One L', 'COREONEL', 1, 1, 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreonel-t1', 'prusa-kb-coreonel', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreonel-t2', 'prusa-kb-coreonel', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreonel-t3', 'prusa-kb-coreonel', 'Lubricate the linear rail', 'days', 0, 90, '', 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-coreonel-t4', 'prusa-kb-coreonel', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 3) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_sets (id, name, model, builtin, enabled, position) VALUES ('prusa-kb-ht90', 'Prusa KB - HT90', 'HT90', 1, 1, 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-ht90-t1', 'prusa-kb-ht90', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-ht90-t2', 'prusa-kb-ht90', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-ht90-t3', 'prusa-kb-ht90', 'Lubricate the linear rail', 'days', 0, 90, '', 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-ht90-t4', 'prusa-kb-ht90', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 3) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_sets (id, name, model, builtin, enabled, position) VALUES ('prusa-kb-xl', 'Prusa KB - XL', 'XL', 1, 1, 3) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-xl-t1', 'prusa-kb-xl', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-xl-t2', 'prusa-kb-xl', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-xl-t3', 'prusa-kb-xl', 'Lubricate the linear rail', 'days', 0, 90, '', 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-xl-t4', 'prusa-kb-xl', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 3) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_sets (id, name, model, builtin, enabled, position) VALUES ('prusa-kb-mk4s', 'Prusa KB - MK4S', 'MK4S', 1, 1, 4) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-mk4s-t1', 'prusa-kb-mk4s', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-mk4s-t2', 'prusa-kb-mk4s', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-mk4s-t3', 'prusa-kb-mk4s', 'Lubricate the linear rail', 'days', 0, 90, '', 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-mk4s-t4', 'prusa-kb-mk4s', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 3) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_sets (id, name, model, builtin, enabled, position) VALUES ('prusa-kb-m3', 'Prusa KB - M3', 'M3', 1, 1, 5) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-m3-t1', 'prusa-kb-m3', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-m3-t2', 'prusa-kb-m3', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-m3-t3', 'prusa-kb-m3', 'Lubricate the linear rail', 'days', 0, 90, '', 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-m3-t4', 'prusa-kb-m3', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 3) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_sets (id, name, model, builtin, enabled, position) VALUES ('prusa-kb-any', 'Prusa KB - Any printer', '*', 1, 1, 6) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-any-t1', 'prusa-kb-any', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-any-t2', 'prusa-kb-any', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-any-t3', 'prusa-kb-any', 'Lubricate the linear rail', 'days', 0, 90, '', 2) ON CONFLICT (id) DO NOTHING;
INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES ('prusa-kb-any-t4', 'prusa-kb-any', 'Check the nozzle, replace when worn', 'hours', 5000, 0, 'hardened steel', 3) ON CONFLICT (id) DO NOTHING;
