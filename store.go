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
        last_done_hours, COALESCE(last_done_at, ''), material
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
			&task.IntervalHours, &task.IntervalDays, &task.LastDoneHours, &task.LastDoneAt, &task.Material); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *store) Task(ctx context.Context, id string) (Task, error) {
	query := fmt.Sprintf(`SELECT id, device_id, name, kind, interval_hours, interval_days,
        last_done_hours, COALESCE(last_done_at, ''), material FROM maintenance_tasks WHERE id = %s`, s.place(1))

	var task Task
	err := s.db.QueryRowContext(ctx, query, id).Scan(&task.ID, &task.DeviceID, &task.Name,
		&task.Kind, &task.IntervalHours, &task.IntervalDays, &task.LastDoneHours, &task.LastDoneAt, &task.Material)

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
        (id, device_id, name, kind, interval_hours, interval_days, last_done_hours, last_done_at, material, created_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind,
        interval_hours = EXCLUDED.interval_hours, interval_days = EXCLUDED.interval_days,
        last_done_hours = EXCLUDED.last_done_hours, last_done_at = EXCLUDED.last_done_at,
        material = EXCLUDED.material`,
		s.place(1), s.place(2), s.place(3), s.place(4), s.place(5), s.place(6), s.place(7), s.place(8), s.place(9), s.place(10))

	_, err := s.db.ExecContext(ctx, query, task.ID, task.DeviceID, task.Name, task.Kind,
		task.IntervalHours, task.IntervalDays, task.LastDoneHours, nullable(task.LastDoneAt), task.Material, nowText())

	return err
}

func (s *store) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM maintenance_tasks WHERE id = %s`, s.place(1))
	_, err := s.db.ExecContext(ctx, query, id)

	return err
}

// PrinterHours is the count the machine shows on its own display, as the
// operator typed it in. The second answer says whether a value was stored.
func (s *store) PrinterHours(ctx context.Context, deviceID string) (float64, bool, error) {
	query := fmt.Sprintf(`SELECT total_hours FROM maintenance_printer_hours WHERE device_id = %s`, s.place(1))

	var hours float64
	err := s.db.QueryRowContext(ctx, query, deviceID).Scan(&hours)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}

	if err != nil {
		return 0, false, err
	}

	return hours, true, nil
}

func (s *store) SetPrinterHours(ctx context.Context, deviceID string, hours float64) error {
	query := fmt.Sprintf(`INSERT INTO maintenance_printer_hours (device_id, total_hours, updated_at)
        VALUES (%s, %s, %s)
        ON CONFLICT (device_id) DO UPDATE SET total_hours = EXCLUDED.total_hours,
        updated_at = EXCLUDED.updated_at`, s.place(1), s.place(2), s.place(3))

	_, err := s.db.ExecContext(ctx, query, deviceID, hours, nowText())

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

func nullable(value string) any {
	if value == "" {
		return nil
	}

	return value
}
