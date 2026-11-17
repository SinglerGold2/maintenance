-- A tool can be reset, because a nozzle gets replaced without the printer's
-- counter going back. The reset stores the hours the tool had at that moment.

ALTER TABLE maintenance_tool_hours ADD COLUMN reset_hours REAL NOT NULL DEFAULT 0;
