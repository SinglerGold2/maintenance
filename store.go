package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
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
	query := fmt.Sprintf(`SELECT tool, hours - reset_hours FROM maintenance_tool_hours WHERE device_id = %s ORDER BY tool`, s.place(1))

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

		if hours < 0 {
			hours = 0
		}

		out[tool] = hours
	}

	return out, rows.Err()
}

// ResetTool starts one tool's counter again at zero, whatever the printer's
// own counter says: a nozzle was replaced, the machine's total did not move.
func (s *store) ResetTool(ctx context.Context, deviceID string, tool int, at float64) error {
	query := fmt.Sprintf(`INSERT INTO maintenance_tool_hours (device_id, tool, hours, reset_hours, updated_at)
        VALUES (%s, %s, %s, %s, %s)
        ON CONFLICT (device_id, tool) DO UPDATE SET hours = EXCLUDED.hours,
        reset_hours = EXCLUDED.reset_hours, updated_at = EXCLUDED.updated_at`,
		s.place(1), s.place(2), s.place(3), s.place(4), s.place(5))

	_, err := s.db.ExecContext(ctx, query, deviceID, tool, at, at, nowText())

	return err
}

// ToolReset is the hours a tool had when it was last reset, and false when it
// never was.
func (s *store) ToolReset(ctx context.Context, deviceID string, tool int) (float64, bool, error) {
	query := fmt.Sprintf(`SELECT reset_hours FROM maintenance_tool_hours WHERE device_id = %s AND tool = %s`,
		s.place(1), s.place(2))

	var reset float64
	err := s.db.QueryRowContext(ctx, query, deviceID, tool).Scan(&reset)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}

	if err != nil {
		return 0, false, err
	}

	return reset, true, nil
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

// Set is a named maintenance schedule. A printer takes its tasks from the set
// it points at, so one printer can run quarterly checks while another follows
// the shipped schedule.
type Set struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Model   string `json:"model"`
	Builtin bool   `json:"builtin"`
	Enabled bool   `json:"enabled"`
	// KBURL is the Knowledge Base article this schedule was read from, when
	// there is one.
	KBURL string    `json:"kb_url,omitempty"`
	Tasks []SetTask `json:"tasks"`
}

// SetTask is one line of a set, before any printer has run it.
type SetTask struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Kind          string  `json:"kind"`
	IntervalHours float64 `json:"interval_hours"`
	IntervalDays  int     `json:"interval_days"`
	Material      string  `json:"material,omitempty"`
}

// Sets lists every schedule with its lines.
func (s *store) Sets(ctx context.Context) ([]Set, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, model, builtin, enabled, kb_url FROM maintenance_sets ORDER BY position, name`)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []Set

	for rows.Next() {
		var (
			set     Set
			builtin int
			enabled int
		)

		if err := rows.Scan(&set.ID, &set.Name, &set.Model, &builtin, &enabled, &set.KBURL); err != nil {
			return nil, err
		}

		set.Builtin = builtin != 0
		set.Enabled = enabled != 0
		out = append(out, set)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		tasks, err := s.SetTasks(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}

		out[i].Tasks = tasks
	}

	return out, nil
}

// SetTasks lists the lines of one set.
func (s *store) SetTasks(ctx context.Context, setID string) ([]SetTask, error) {
	query := fmt.Sprintf(`SELECT id, name, kind, interval_hours, interval_days, material FROM maintenance_set_tasks WHERE set_id = %s ORDER BY position`, s.place(1))

	rows, err := s.db.QueryContext(ctx, query, setID)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var out []SetTask

	for rows.Next() {
		var task SetTask
		if err := rows.Scan(&task.ID, &task.Name, &task.Kind, &task.IntervalHours, &task.IntervalDays, &task.Material); err != nil {
			return nil, err
		}

		out = append(out, task)
	}

	return out, rows.Err()
}

// CreateSet writes a new schedule, optionally a copy of another one.
func (s *store) CreateSet(ctx context.Context, name, model, copyOf string) (Set, error) {
	set := Set{ID: newID(), Name: name, Model: model, Enabled: true, Tasks: []SetTask{}}

	query := fmt.Sprintf(`INSERT INTO maintenance_sets (id, name, model, builtin, enabled, kb_url, position) VALUES (%s, %s, %s, 0, 1, '', (SELECT COALESCE(MAX(position), 0) + 1 FROM maintenance_sets))`,
		s.place(1), s.place(2), s.place(3))

	if _, err := s.db.ExecContext(ctx, query, set.ID, set.Name, set.Model); err != nil {
		return Set{}, err
	}

	if copyOf != "" {
		tasks, err := s.SetTasks(ctx, copyOf)
		if err != nil {
			return Set{}, err
		}

		for _, task := range tasks {
			if _, err := s.AddSetTask(ctx, set.ID, SetTask{Name: task.Name, Kind: task.Kind, IntervalHours: task.IntervalHours, IntervalDays: task.IntervalDays, Material: task.Material}); err != nil {
				return Set{}, err
			}
		}

		if set.Tasks, err = s.SetTasks(ctx, set.ID); err != nil {
			return Set{}, err
		}
	}

	return set, nil
}

// UpdateSet renames a schedule, or switches it on and off.
func (s *store) UpdateSet(ctx context.Context, id, name, model string, enabled bool) error {
	query := fmt.Sprintf(`UPDATE maintenance_sets SET name = %s, model = %s, enabled = %s WHERE id = %s`,
		s.place(1), s.place(2), s.place(3), s.place(4))

	_, err := s.db.ExecContext(ctx, query, name, model, enabled, id)

	return err
}

// SetIsBuiltin reports whether a schedule shipped with this module.
func (s *store) SetIsBuiltin(ctx context.Context, id string) (bool, error) {
	query := fmt.Sprintf(`SELECT builtin FROM maintenance_sets WHERE id = %s`, s.place(1))

	var builtin int
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&builtin); err != nil {
		return false, err
	}

	return builtin != 0, nil
}

// DeleteSet removes a schedule and its lines.
func (s *store) DeleteSet(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM maintenance_sets WHERE id = %s`, s.place(1))
	_, err := s.db.ExecContext(ctx, query, id)

	return err
}

// SetInUse counts the printers pointed at a schedule.
func (s *store) SetInUse(ctx context.Context, id string) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM maintenance_printer_sets WHERE set_id = %s`, s.place(1))

	var count int
	err := s.db.QueryRowContext(ctx, query, id).Scan(&count)

	return count, err
}

// AddSetTask appends a line to a schedule.
func (s *store) AddSetTask(ctx context.Context, setID string, task SetTask) (SetTask, error) {
	task.ID = newID()

	query := fmt.Sprintf(`INSERT INTO maintenance_set_tasks (id, set_id, name, kind, interval_hours, interval_days, material, position) VALUES (%s, %s, %s, %s, %s, %s, %s, (SELECT COALESCE(MAX(position), 0) + 1 FROM maintenance_set_tasks WHERE set_id = %s))`,
		s.place(1), s.place(2), s.place(3), s.place(4), s.place(5), s.place(6), s.place(7), s.place(8))

	_, err := s.db.ExecContext(ctx, query, task.ID, setID, task.Name, task.Kind, task.IntervalHours, task.IntervalDays, task.Material, setID)
	if err != nil {
		return SetTask{}, err
	}

	return task, nil
}

// UpdateSetTask changes one line of a schedule.
func (s *store) UpdateSetTask(ctx context.Context, setID string, task SetTask) error {
	query := fmt.Sprintf(`UPDATE maintenance_set_tasks SET name = %s, kind = %s, interval_hours = %s, interval_days = %s, material = %s WHERE id = %s AND set_id = %s`,
		s.place(1), s.place(2), s.place(3), s.place(4), s.place(5), s.place(6), s.place(7))

	_, err := s.db.ExecContext(ctx, query, task.Name, task.Kind, task.IntervalHours, task.IntervalDays, task.Material, task.ID, setID)

	return err
}

// DeleteSetTask removes one line of a schedule.
func (s *store) DeleteSetTask(ctx context.Context, setID, taskID string) error {
	query := fmt.Sprintf(`DELETE FROM maintenance_set_tasks WHERE id = %s AND set_id = %s`, s.place(1), s.place(2))
	_, err := s.db.ExecContext(ctx, query, taskID, setID)

	return err
}

// PrinterSet is the schedule a printer follows, or empty when it follows the
// one its model brings.
func (s *store) PrinterSet(ctx context.Context, deviceID string) (string, error) {
	query := fmt.Sprintf(`SELECT set_id FROM maintenance_printer_sets WHERE device_id = %s`, s.place(1))

	var setID string
	err := s.db.QueryRowContext(ctx, query, deviceID).Scan(&setID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return setID, err
}

// AssignPrinterSet points a printer at a schedule.
func (s *store) AssignPrinterSet(ctx context.Context, deviceID, setID string) error {
	query := fmt.Sprintf(`INSERT INTO maintenance_printer_sets (device_id, set_id) VALUES (%s, %s) ON CONFLICT (device_id) DO UPDATE SET set_id = EXCLUDED.set_id`,
		s.place(1), s.place(2))

	_, err := s.db.ExecContext(ctx, query, deviceID, setID)

	return err
}

// DefaultSet is the schedule a model brings: its own when one is enabled, the
// one for any printer otherwise.
func (s *store) DefaultSet(ctx context.Context, model string) (string, error) {
	upper := strings.ToUpper(strings.ReplaceAll(model, " ", ""))

	for _, candidate := range []string{upper, "*"} {
		query := fmt.Sprintf(`SELECT id FROM maintenance_sets WHERE enabled = %s AND model = %s ORDER BY position LIMIT 1`, s.place(1), s.place(2))

		var id string
		err := s.db.QueryRowContext(ctx, query, 1, candidate).Scan(&id)
		if err == nil {
			return id, nil
		}

		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}

	return "", nil
}

// SyncSetTasks makes a printer's builtin tasks match the schedule it follows:
// missing lines are created, intervals follow the set, and a line removed from
// the set is removed from the printer. Its own tasks are left alone.
func (s *store) SyncSetTasks(ctx context.Context, deviceID, setID string, now time.Time) error {
	lines, err := s.SetTasks(ctx, setID)
	if err != nil {
		return err
	}

	existing, err := s.Tasks(ctx, deviceID)
	if err != nil {
		return err
	}

	byName := make(map[string]Task, len(existing))
	for _, task := range existing {
		if task.Builtin {
			byName[task.Name] = task
		}
	}

	want := make(map[string]bool, len(lines))
	stamp := now.Format(time.RFC3339)

	for _, line := range lines {
		want[line.Name] = true

		current, found := byName[line.Name]
		if !found {
			// A set names a nozzle line by its material; the printer's own task
			// only carries the interval that material gives.
			kind := line.Kind
			if kind == "nozzle" {
				kind = "hours"
				if line.IntervalDays > 0 {
					kind = "days"
				}
			}

			insert := Task{
				ID:            newID(),
				DeviceID:      deviceID,
				Name:          line.Name,
				Kind:          kind,
				IntervalHours: line.IntervalHours,
				IntervalDays:  line.IntervalDays,
				Material:      line.Material,
				LastDoneAt:    stamp,
				Builtin:       true,
			}

			if err := s.Upsert(ctx, insert); err != nil {
				return err
			}

			continue
		}

		if current.Kind == line.Kind && current.Kind != "nozzle" && current.IntervalHours == line.IntervalHours &&
			current.IntervalDays == line.IntervalDays && current.Material == line.Material {
			continue
		}

		current.Kind = line.Kind
		if current.Kind == "nozzle" {
			current.Kind = "hours"
			if line.IntervalDays > 0 {
				current.Kind = "days"
			}
		}

		current.IntervalHours = line.IntervalHours
		current.IntervalDays = line.IntervalDays
		current.Material = line.Material

		if err := s.Upsert(ctx, current); err != nil {
			return err
		}
	}

	for name, task := range byName {
		if want[name] {
			continue
		}

		if err := s.Delete(ctx, task.ID); err != nil {
			return err
		}
	}

	return nil
}
