package main

import (
	"context"
	"database/sql"
	"fmt"
)

// store keeps the maintenance tasks in the module's own database. The two
// engines take different placeholders, so the statements are built from one
// helper rather than written twice.
type store struct {
	db       *sql.DB
	postgres bool
}

// place is the placeholder of the n-th (one-based) argument.
func (s *store) place(n int) string {
	if s.postgres {
		return fmt.Sprintf("$%d", n)
	}

	return "?"
}

func (s *store) Tasks(ctx context.Context, deviceID string) ([]Task, error) {
	query := fmt.Sprintf(`SELECT id, device_id, name, kind, interval_hours, interval_days,
        last_done_hours, COALESCE(last_done_at, ''), material, builtin
        FROM maintenance_tasks WHERE device_id = %s ORDER BY created_at, name`, s.place(1))

	rows, err := s.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.DeviceID, &task.Name, &task.Kind,
			&task.IntervalHours, &task.IntervalDays, &task.LastDoneHours, &task.LastDoneAt, &task.Material, &task.Builtin); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *store) Task(ctx context.Context, id string) (Task, error) {
	query := fmt.Sprintf(`SELECT id, device_id, name, kind, interval_hours, interval_days,
        last_done_hours, COALESCE(last_done_at, ''), material, builtin FROM maintenance_tasks WHERE id = %s`, s.place(1))

	var task Task
	err := s.db.QueryRowContext(ctx, query, id).Scan(&task.ID, &task.DeviceID, &task.Name,
		&task.Kind, &task.IntervalHours, &task.IntervalDays, &task.LastDoneHours, &task.LastDoneAt, &task.Material, &task.Builtin)

	return task, err
}

func (s *store) Insert(ctx context.Context, tasks []Task) error {
	for _, task := range tasks {
		if err := s.Upsert(ctx, task); err != nil {
			return err
		}
	}

	return nil
}

func (s *store) Upsert(ctx context.Context, task Task) error {
	query := fmt.Sprintf(`INSERT INTO maintenance_tasks
        (id, device_id, name, kind, interval_hours, interval_days, last_done_hours, last_done_at, material, builtin, created_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind,
        interval_hours = EXCLUDED.interval_hours, interval_days = EXCLUDED.interval_days,
        last_done_hours = EXCLUDED.last_done_hours, last_done_at = EXCLUDED.last_done_at,
        material = EXCLUDED.material`,
		s.place(1), s.place(2), s.place(3), s.place(4), s.place(5), s.place(6), s.place(7), s.place(8), s.place(9), s.place(10), s.place(11))

	_, err := s.db.ExecContext(ctx, query, task.ID, task.DeviceID, task.Name, task.Kind,
		task.IntervalHours, task.IntervalDays, task.LastDoneHours, nullable(task.LastDoneAt), task.Material, task.Builtin, nowText())

	return err
}

func (s *store) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM maintenance_tasks WHERE id = %s`, s.place(1))
	_, err := s.db.ExecContext(ctx, query, id)

	return err
}

// PrinterCounts is the count the operator read off the machine and the distance
// this module keeps between that count and the hours it has tracked. A negative
// distance marks a row written before the distance was kept: the module rebuilds
// it from the count and the hours it has tracked.
func (s *store) PrinterCounts(ctx context.Context, deviceID string) (float64, float64, bool, error) {
	query := fmt.Sprintf(`SELECT total_hours, offset_hours FROM maintenance_printer_hours WHERE device_id = %s`, s.place(1))

	var total, offset float64
	err := s.db.QueryRowContext(ctx, query, deviceID).Scan(&total, &offset)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	}

	if err != nil {
		return 0, 0, false, err
	}

	return total, offset, true, nil
}

// SetPrinterOffset rewrites the distance alone, keeping the count.
func (s *store) SetPrinterOffset(ctx context.Context, deviceID string, offset float64) error {
	query := fmt.Sprintf(`UPDATE maintenance_printer_hours SET offset_hours = %s, updated_at = %s WHERE device_id = %s`,
		s.place(1), s.place(2), s.place(3))

	_, err := s.db.ExecContext(ctx, query, offset, nowText(), deviceID)

	return err
}

// SetPrinterHours stores the distance between the value the operator read off
// the machine and the hours this server has tracked so far.
func (s *store) SetPrinterHours(ctx context.Context, deviceID string, printerHours, trackedHours float64) error {
	query := fmt.Sprintf(`INSERT INTO maintenance_printer_hours (device_id, total_hours, offset_hours, updated_at)
        VALUES (%s, %s, %s, %s)
        ON CONFLICT (device_id) DO UPDATE SET total_hours = EXCLUDED.total_hours,
        offset_hours = EXCLUDED.offset_hours, updated_at = EXCLUDED.updated_at`,
		s.place(1), s.place(2), s.place(3), s.place(4))

	_, err := s.db.ExecContext(ctx, query, deviceID, printerHours, printerHours-trackedHours, nowText())

	return err
}

// TemplateTask is one line of the schedule that every printer starts with.
type TemplateTask struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Kind          string  `json:"kind"`
	IntervalHours float64 `json:"interval_hours"`
	IntervalDays  int     `json:"interval_days"`
	Material      string  `json:"material,omitempty"`
	Position      int     `json:"position"`
}

func (s *store) Template(ctx context.Context) ([]TemplateTask, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, kind, interval_hours, interval_days,
        material, position FROM maintenance_template ORDER BY position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TemplateTask

	for rows.Next() {
		var task TemplateTask
		if err := rows.Scan(&task.ID, &task.Name, &task.Kind, &task.IntervalHours,
			&task.IntervalDays, &task.Material, &task.Position); err != nil {
			return nil, err
		}

		out = append(out, task)
	}

	return out, rows.Err()
}

// UpdateTemplate writes one line of the schedule and carries it to every
// built-in task of that name, keeping each printer's own counter.
func (s *store) UpdateTemplate(ctx context.Context, task TemplateTask) error {
	update := fmt.Sprintf(`UPDATE maintenance_template SET interval_hours = %s, interval_days = %s,
        material = %s WHERE id = %s`, s.place(1), s.place(2), s.place(3), s.place(4))

	if _, err := s.db.ExecContext(ctx, update, task.IntervalHours, task.IntervalDays, task.Material, task.ID); err != nil {
		return err
	}

	apply := fmt.Sprintf(`UPDATE maintenance_tasks SET interval_hours = %s, interval_days = %s,
        material = %s WHERE builtin = %s AND name = %s`,
		s.place(1), s.place(2), s.place(3), s.place(4), s.place(5))

	_, err := s.db.ExecContext(ctx, apply, task.IntervalHours, task.IntervalDays, task.Material, true, task.Name)

	return err
}

// MarkSeeded answers whether this call is the one that seeds the printer's
// schedule. The first caller wins and every later one is told no.
func (s *store) MarkSeeded(ctx context.Context, deviceID string) (bool, error) {
	query := fmt.Sprintf(`INSERT INTO maintenance_seeded (device_id, seeded_at) VALUES (%s, %s)
        ON CONFLICT (device_id) DO NOTHING`, s.place(1), s.place(2))

	result, err := s.db.ExecContext(ctx, query, deviceID, nowText())
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows == 1, nil
}

// JobSeen answers whether this print was already split across the tools.
func (s *store) JobSeen(ctx context.Context, deviceID, jobID string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM maintenance_jobs_seen WHERE device_id = %s AND job_id = %s`,
		s.place(1), s.place(2))

	var one int
	err := s.db.QueryRowContext(ctx, query, deviceID, jobID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *store) MarkJobSeen(ctx context.Context, deviceID, jobID string) error {
	query := fmt.Sprintf(`INSERT INTO maintenance_jobs_seen (device_id, job_id, seen_at)
        VALUES (%s, %s, %s) ON CONFLICT (device_id, job_id) DO NOTHING`,
		s.place(1), s.place(2), s.place(3))

	_, err := s.db.ExecContext(ctx, query, deviceID, jobID, nowText())

	return err
}

// ToolHours is the hours each tool has run on one printer.
func (s *store) ToolHours(ctx context.Context, deviceID string) (map[int]float64, error) {
	query := fmt.Sprintf(`SELECT tool, hours FROM maintenance_tool_hours WHERE device_id = %s ORDER BY tool`, s.place(1))

	rows, err := s.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int]float64{}

	for rows.Next() {
		var (
			tool  int
			hours float64
		)

		if err := rows.Scan(&tool, &hours); err != nil {
			return nil, err
		}

		out[tool] = hours
	}

	return out, rows.Err()
}

func (s *store) AddToolHours(ctx context.Context, deviceID string, tool int, hours float64) error {
	query := fmt.Sprintf(`INSERT INTO maintenance_tool_hours (device_id, tool, hours, updated_at)
        VALUES (%s, %s, %s, %s)
        ON CONFLICT (device_id, tool) DO UPDATE SET hours = maintenance_tool_hours.hours + EXCLUDED.hours,
        updated_at = EXCLUDED.updated_at`, s.place(1), s.place(2), s.place(3), s.place(4))

	_, err := s.db.ExecContext(ctx, query, deviceID, tool, hours, nowText())

	return err
}

func nullable(value string) any {
	if value == "" {
		return nil
	}

	return value
}
