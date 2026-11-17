-- The MK3 family has a category page rather than one article, so it gets its
-- own schedule with that link, and the four lines the Knowledge Base lists.
INSERT INTO maintenance_sets (id, name, model, builtin, enabled, kb_url, position)
VALUES ('prusa-kb-mk3', 'Prusa KB - MK3', 'MK3', 1, 1, 'https://help.prusa3d.com/category/printer-maintenance_247', 3)
ON CONFLICT (id) DO NOTHING;

INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position)
VALUES
 ('prusa-kb-mk3-t1', 'prusa-kb-mk3', 'Check axes, belts, fans and extruder', 'hours', 200, 0, '', 0),
 ('prusa-kb-mk3-t2', 'prusa-kb-mk3', 'Inspect the electrical connectors', 'hours', 600, 0, '', 1),
 ('prusa-kb-mk3-t3', 'prusa-kb-mk3', 'Lubricate the linear rail', 'days', 0, 90, '', 2),
 ('prusa-kb-mk3-t4', 'prusa-kb-mk3', 'Check the nozzle, replace when worn', 'nozzle', 1500, 0, 'hardened steel', 3)
ON CONFLICT (id) DO NOTHING;
