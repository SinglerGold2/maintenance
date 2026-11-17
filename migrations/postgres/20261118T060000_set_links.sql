-- Every shipped schedule points at the Prusa Knowledge Base article it was
-- read from, so an operator can check the intervals against the source.
ALTER TABLE maintenance_sets ADD COLUMN kb_url TEXT NOT NULL DEFAULT '';

UPDATE maintenance_sets SET kb_url = 'https://help.prusa3d.com/article/regular-printer-maintenance-core-one_829710' WHERE id = 'prusa-kb-coreone';
UPDATE maintenance_sets SET kb_url = 'https://help.prusa3d.com/article/regular-printer-maintenance-core-one-l_970074' WHERE id = 'prusa-kb-coreonel';
UPDATE maintenance_sets SET kb_url = 'https://help.prusa3d.com/article/regular-printer-maintenance-ht90_725956' WHERE id = 'prusa-kb-ht90';
UPDATE maintenance_sets SET kb_url = 'https://help.prusa3d.com/article/regular-printer-maintenance-xl_401237' WHERE id = 'prusa-kb-xl';
UPDATE maintenance_sets SET kb_url = 'https://help.prusa3d.com/article/regular-printer-maintenance-mk4-s-mk3-9-s_419000' WHERE id = 'prusa-kb-mk4s';
