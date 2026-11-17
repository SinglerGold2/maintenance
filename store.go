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
        last_done_hours, COALESCE(last_done_at, '')
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
			&task.IntervalHours, &task.IntervalDays, &task.LastDoneHours, &task.LastDoneAt); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *store) Task(ctx context.Context, id string) (Task, error) {
	query := fmt.Sprintf(`SELECT id, device_id, name, kind, interval_hours, interval_days,
        last_done_hours, COALESCE(last_done_at, '') FROM maintenance_tasks WHERE id = %s`, s.place(1))

	var task Task
	err := s.db.QueryRowContext(ctx, query, id).Scan(&task.ID, &task.DeviceID, &task.Name,
		&task.Kind, &task.IntervalHours, &task.IntervalDays, &task.LastDoneHours, &task.LastDoneAt)

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
        (id, device_id, name, kind, interval_hours, interval_days, last_done_hours, last_done_at, created_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, kind = EXCLUDED.kind,
        interval_hours = EXCLUDED.interval_hours, interval_days = EXCLUDED.interval_days,
        last_done_hours = EXCLUDED.last_done_hours, last_done_at = EXCLUDED.last_done_at`,
		s.place(1), s.place(2), s.place(3), s.place(4), s.place(5), s.place(6), s.place(7), s.place(8), s.place(9))

	_, err := s.db.ExecContext(ctx, query, task.ID, task.DeviceID, task.Name, task.Kind,
		task.IntervalHours, task.IntervalDays, task.LastDoneHours, nullable(task.LastDoneAt), nowText())

	return err
}

func (s *store) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM maintenance_tasks WHERE id = %s`, s.place(1))
	_, err := s.db.ExecContext(ctx, query, id)

	return err
}

func nullable(value string) any {
	if value == "" {
		return nil
	}

	return value
}
