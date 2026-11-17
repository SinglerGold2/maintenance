// Command maintenance counts the print hours of each printer and reminds the
// operator when a maintenance task is due.
//
// It keeps one row per task in its own database, reads the print history the
// host holds, and serves one page in the Connect web app.
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/prusa3d/prusa-connect/sdk/deviceapi"
	"github.com/prusa3d/prusa-connect/sdk/extension"
	"github.com/prusa3d/prusa-connect/sdk/moduledb"
	"github.com/prusa3d/prusa-connect/sdk/routesapi"
	"github.com/prusa3d/prusa-connect/sdk/wire"
	"github.com/prusa3d/prusa-connect/sdk/wirerpc"
)

//go:embed module.yaml
var manifestYAML []byte

//go:embed openapi.json
var openAPI []byte

//go:embed migrations/sqlite/* migrations/postgres/*
var migrations embed.FS

//nolint:gochecknoglobals // a fixed description of the embedded files.
var schema = moduledb.Schema{Migrations: migrations}

//nolint:gochecknoglobals // read once from the embedded manifest.
var (
	name    = manifestScalar(manifestYAML, "name")
	release = extension.Version(manifestScalar(manifestYAML, "version"))
)

const defaultAlertPercent = 90

// Settings is what an operator writes under this instance's config block.
type Settings struct {
	// AlertPercent is the share of an interval at which a task reads as due soon.
	AlertPercent int `json:"alert_percent"`
}

func manifestScalar(document []byte, key string) string {
	for _, line := range strings.Split(string(document), "\n") {
		value, found := strings.CutPrefix(line, key+":")
		if !found {
			continue
		}

		if comment := strings.Index(value, " #"); comment >= 0 {
			value = value[:comment]
		}

		return strings.Trim(strings.TrimSpace(value), `"'`)
	}

	return ""
}

// Task is one maintenance job of one printer.
type Task struct {
	ID            string  `json:"id"`
	DeviceID      string  `json:"device_id"`
	Name          string  `json:"name"`
	Kind          string  `json:"kind"`
	IntervalHours float64 `json:"interval_hours"`
	IntervalDays  int     `json:"interval_days"`
	LastDoneHours float64 `json:"last_done_hours"`
	LastDoneAt    string  `json:"last_done_at,omitempty"`
	Material      string  `json:"material,omitempty"`
}

// TaskView is a task with what the counters say about it.
type TaskView struct {
	Task
	Elapsed   float64 `json:"elapsed"`
	Remaining float64 `json:"remaining"`
	Percent   float64 `json:"percent"`
	Status    string  `json:"status"`
}

// DeviceView is one printer with its hours and tasks.
type DeviceView struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Model  string  `json:"model"`
	Hours  float64 `json:"hours"`
	Prints int     `json:"prints"`
	// PrinterHours is what the machine's own display said, when the operator
	// typed it in.
	PrinterHours *float64   `json:"printer_hours,omitempty"`
	Tasks        []TaskView `json:"tasks"`
}

// module is the state the routes read.
type module struct {
	mu       sync.Mutex
	db       *sql.DB
	postgres bool
	devices  *deviceapi.Client
	alert    int

	cacheMu sync.Mutex
	cache   map[string]hoursEntry
}

type hoursEntry struct {
	hours      float64
	prints     int
	calculated time.Time
}

const hoursTTL = 5 * time.Minute

var errNoDatabase = errors.New("the operator granted this module no database. Approve its needs.database request")

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	line, err := wire.Handshake{
		Protocol: wire.ProtocolMajor,
		MinMinor: wire.MinMinor,
		MaxMinor: wire.MaxMinor,
		Address:  wire.StdioAddress,
	}.Line()
	if err != nil {
		return fmt.Errorf("render the handshake: %w", err)
	}

	if _, err = os.Stdout.WriteString(line); err != nil {
		return fmt.Errorf("write the handshake: %w", err)
	}

	conn, err := wirerpc.NewConn(stdio{}, wirerpc.Options{
		Role:    wirerpc.RoleChild,
		Methods: methods(),
	})
	if err != nil {
		return fmt.Errorf("build the session: %w", err)
	}

	state.mu.Lock()
	state.devices = deviceapi.NewClient(conn)
	state.mu.Unlock()

	fmt.Fprintf(os.Stderr, "%s %s is serving on stdio as pid %d\n", name, release, os.Getpid())

	if err = conn.Serve(ctx); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

//nolint:gochecknoglobals // the one piece of state this module has.
var state = &module{alert: defaultAlertPercent, cache: map[string]hoursEntry{}}

func methods() wirerpc.Methods {
	return wirerpc.Methods{
		wirerpc.MethodInfo: func(context.Context, wirerpc.Call) (any, error) {
			return routesapi.Describe(
				wirerpc.Info{
					Name:     name,
					Version:  release,
					Contract: wirerpc.Contract,
					Serves:   []wirerpc.Method{wirerpc.MethodConfigure},
				}, openAPI)
		},
		wirerpc.MethodConfigure: wirerpc.ConfigureGranted(func(ctx context.Context, sent Settings, grants wirerpc.Grants) error {
			if sent.AlertPercent < 1 || sent.AlertPercent > 99 {
				sent.AlertPercent = defaultAlertPercent
			}

			if grants.Database == nil {
				return errNoDatabase
			}

			db, _, err := moduledb.Open(ctx, *grants.Database, schema)
			if err != nil {
				return fmt.Errorf("%w (the database is %s)", err, grants.Database)
			}

			state.mu.Lock()
			state.db = db
			state.postgres = grants.Database.Engine == wirerpc.DatabasePostgres
			state.alert = sent.AlertPercent
			state.mu.Unlock()

			return nil
		}),
		wirerpc.MethodShutdown: func(context.Context, wirerpc.Call) (any, error) {
			fmt.Fprintf(os.Stderr, "%s was asked to stop\n", name)

			return nil, nil
		},
		routesapi.MethodServe: routesapi.Serve(routes()),
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /overview", handleOverview)
	mux.HandleFunc("POST /tasks", handleSaveTask)
	mux.HandleFunc("POST /tasks/done", handleDone)
	mux.HandleFunc("DELETE /tasks", handleDelete)
	mux.HandleFunc("POST /devices/hours", handlePrinterHours)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		refuse(w, r, RouteNotFound.
			New("this module serves no route at "+r.URL.Path).
			With(faultString("path", r.URL.Path)))
	})

	return mux
}

func handleOverview(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	devices, err := state.listDevices(r.Context())
	if err != nil {
		refuse(w, r, HistoryFailed.New("the fleet could not be read").With(faultString("reason", err.Error())))
		return
	}

	views := make([]DeviceView, 0, len(devices))

	for _, device := range devices {
		hours, prints, hoursErr := state.hoursOf(r.Context(), device.ID)
		if hoursErr != nil {
			refuse(w, r, HistoryFailed.New("the print history could not be read").With(faultString("reason", hoursErr.Error())))
			return
		}

		tasks, taskErr := st.Tasks(r.Context(), device.ID)
		if taskErr != nil {
			refuse(w, r, StoreFailed.New("the tasks could not be read").With(faultString("reason", taskErr.Error())))
			return
		}

		if len(tasks) == 0 {
			first, seedErr := st.MarkSeeded(r.Context(), device.ID)
			if seedErr != nil {
				refuse(w, r, StoreFailed.New("the seed marker could not be written").With(faultString("reason", seedErr.Error())))
				return
			}

			if first {
				tasks = defaultsFor(device.Model, time.Now().UTC())
				if seedErr = st.Insert(r.Context(), tasks); seedErr != nil {
					refuse(w, r, StoreFailed.New("the default tasks could not be written").With(faultString("reason", seedErr.Error())))
					return
				}
			}
		}

		printerHours, known, hoursErr := st.PrinterHours(r.Context(), device.ID)
		if hoursErr != nil {
			refuse(w, r, StoreFailed.New("the printer's own hours could not be read").With(faultString("reason", hoursErr.Error())))
			return
		}

		var typed *float64
		if known {
			typed = &printerHours
		}

		views = append(views, DeviceView{
			ID:           device.ID,
			Name:         device.Name,
			Model:        device.Model,
			Hours:        round1(hours),
			Prints:       prints,
			PrinterHours: typed,
			Tasks:        state.view(tasks, hours, time.Now()),
		})
	}

	answer(w, http.StatusOK, map[string]any{"devices": views, "alert_percent": state.alertPercent()})
}

type saveRequest struct {
	ID            string   `json:"id"`
	DeviceID      string   `json:"device_id"`
	Name          string   `json:"name"`
	Kind          string   `json:"kind"`
	IntervalHours float64  `json:"interval_hours"`
	IntervalDays  int      `json:"interval_days"`
	Material      string   `json:"material"`
	LastDoneHours *float64 `json:"last_done_hours"`
}

func handleSaveTask(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	var sent saveRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&sent); decodeErr != nil {
		refuse(w, r, InvalidInput.New("the request body is not a task"))
		return
	}

	if sent.DeviceID == "" || sent.Name == "" || (sent.Kind != "hours" && sent.Kind != "days") {
		refuse(w, r, InvalidInput.New("device, name and kind are required").With(faultString("kind", sent.Kind)))
		return
	}

	if sent.Kind == "hours" && sent.IntervalHours <= 0 {
		refuse(w, r, InvalidInput.New("an hours task needs a positive interval"))
		return
	}

	if sent.Kind == "days" && sent.IntervalDays <= 0 {
		refuse(w, r, InvalidInput.New("a days task needs a positive interval"))
		return
	}

	task := Task{
		ID:            sent.ID,
		DeviceID:      sent.DeviceID,
		Name:          strings.TrimSpace(sent.Name),
		Kind:          sent.Kind,
		IntervalHours: sent.IntervalHours,
		IntervalDays:  sent.IntervalDays,
		Material:      strings.ToLower(strings.TrimSpace(sent.Material)),
	}

	if sent.LastDoneHours != nil {
		task.LastDoneHours = *sent.LastDoneHours
	}

	if task.ID == "" {
		task.ID = newID()
		task.LastDoneAt = time.Now().UTC().Format(time.RFC3339)
	}

	if saveErr := st.Upsert(r.Context(), task); saveErr != nil {
		refuse(w, r, StoreFailed.New("the task could not be written").With(faultString("reason", saveErr.Error())))
		return
	}

	found, foundErr := st.Task(r.Context(), task.ID)
	if foundErr != nil {
		refuse(w, r, StoreFailed.New("the task could not be read back").With(faultString("reason", foundErr.Error())))
		return
	}

	answer(w, http.StatusOK, found)
}

func handleDone(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	var sent struct {
		ID string `json:"id"`
	}
	if decodeErr := json.NewDecoder(r.Body).Decode(&sent); decodeErr != nil || sent.ID == "" {
		refuse(w, r, InvalidInput.New("the request body names no task"))
		return
	}

	task, foundErr := st.Task(r.Context(), sent.ID)
	if foundErr != nil {
		refuse(w, r, NoTask.New("no task carries that id").With(faultString("id", sent.ID)))
		return
	}

	hours, _, hoursErr := state.hoursOf(r.Context(), task.DeviceID)
	if hoursErr != nil {
		refuse(w, r, HistoryFailed.New("the print history could not be read").With(faultString("reason", hoursErr.Error())))
		return
	}

	task.LastDoneHours = hours
	task.LastDoneAt = time.Now().UTC().Format(time.RFC3339)

	if saveErr := st.Upsert(r.Context(), task); saveErr != nil {
		refuse(w, r, StoreFailed.New("the task could not be written").With(faultString("reason", saveErr.Error())))
		return
	}

	answer(w, http.StatusOK, task)
}

func handlePrinterHours(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	var sent struct {
		DeviceID string  `json:"device_id"`
		Hours    float64 `json:"hours"`
	}

	if decodeErr := json.NewDecoder(r.Body).Decode(&sent); decodeErr != nil || sent.DeviceID == "" {
		refuse(w, r, InvalidInput.New("the request names no printer"))
		return
	}

	if sent.Hours < 0 {
		refuse(w, r, InvalidInput.New("hours cannot be negative"))
		return
	}

	if storeErr := st.SetPrinterHours(r.Context(), sent.DeviceID, sent.Hours); storeErr != nil {
		refuse(w, r, StoreFailed.New("the printer's own hours could not be written").With(faultString("reason", storeErr.Error())))
		return
	}

	answer(w, http.StatusOK, map[string]any{"device_id": sent.DeviceID, "hours": sent.Hours})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		refuse(w, r, InvalidInput.New("the query names no task"))
		return
	}

	if deleteErr := st.Delete(r.Context(), id); deleteErr != nil {
		refuse(w, r, StoreFailed.New("the task could not be removed").With(faultString("reason", deleteErr.Error())))
		return
	}

	answer(w, http.StatusOK, map[string]any{"deleted": id})
}

func (m *module) store() (*store, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.db == nil {
		return nil, errNoDatabase
	}

	return &store{db: m.db, postgres: m.postgres}, nil
}

func (m *module) alertPercent() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.alert
}

func (m *module) listDevices(ctx context.Context) ([]deviceapi.Device, error) {
	m.mu.Lock()
	client := m.devices
	m.mu.Unlock()

	if client == nil {
		return nil, errors.New("no device client")
	}

	page, err := client.List(ctx, deviceapi.ListParams{Limit: 100})
	if err != nil {
		return nil, err
	}

	return page.Items, nil
}

// hoursOf sums the length of every ended print of one device. The result is
// held for a few minutes, because a page load asks for it once per device.
func (m *module) hoursOf(ctx context.Context, deviceID string) (float64, int, error) {
	m.cacheMu.Lock()
	cached, found := m.cache[deviceID]
	m.cacheMu.Unlock()

	if found && time.Since(cached.calculated) < hoursTTL {
		return cached.hours, cached.prints, nil
	}

	m.mu.Lock()
	client := m.devices
	m.mu.Unlock()

	if client == nil {
		return 0, 0, errors.New("no device client")
	}

	var (
		total  float64
		prints int
		cursor string
	)

	for page := 0; page < 50; page++ {
		history, err := client.JobHistory(ctx, deviceapi.JobHistoryParams{ID: deviceID, Limit: 200, Cursor: cursor})
		if err != nil {
			return 0, 0, err
		}

		for _, job := range history.Items {
			if job.Ended == nil || job.Started.IsZero() {
				continue
			}

			length := job.Ended.Sub(job.Started).Hours()
			if length <= 0 || length > 1000 {
				continue
			}

			total += length
			prints++
		}

		if history.Cursor == "" || history.Cursor == cursor {
			break
		}

		cursor = history.Cursor
	}

	m.cacheMu.Lock()
	m.cache[deviceID] = hoursEntry{hours: total, prints: prints, calculated: time.Now()}
	m.cacheMu.Unlock()

	return total, prints, nil
}

func (m *module) view(tasks []Task, hours float64, now time.Time) []TaskView {
	alert := float64(m.alertPercent()) / 100
	views := make([]TaskView, 0, len(tasks))

	for _, task := range tasks {
		var elapsed, interval float64

		switch task.Kind {
		case "hours":
			interval = task.IntervalHours
			elapsed = hours - task.LastDoneHours
			if elapsed < 0 {
				elapsed = 0
			}
		case "days":
			interval = float64(task.IntervalDays) * 24
			last := now
			if task.LastDoneAt != "" {
				if parsed, err := time.Parse(time.RFC3339, task.LastDoneAt); err == nil {
					last = parsed
				}
			}
			elapsed = now.Sub(last).Hours()
		}

		percent := 0.0
		if interval > 0 {
			percent = elapsed / interval * 100
		}

		status := "ok"

		switch {
		case percent >= 100:
			status = "due"
		case percent >= alert*100:
			status = "soon"
		}

		views = append(views, TaskView{
			Task:      task,
			Elapsed:   round1(elapsed),
			Remaining: round1(interval - elapsed),
			Percent:   round1(percent),
			Status:    status,
		})
	}

	return views
}

// defaultsFor is the maintenance schedule Prusa publishes for a model. The
// numbers are from the Prusa Knowledge Base article on regular printer
// maintenance.
func defaultsFor(model string, now time.Time) []Task {
	upper := strings.ToUpper(strings.ReplaceAll(model, " ", ""))
	stamp := now.Format(time.RFC3339)

	if !strings.Contains(upper, "COREONE") {
		return []Task{
			{ID: newID(), Name: "Inspect the machine", Kind: "hours", IntervalHours: 200, LastDoneAt: stamp},
		}
	}

	return []Task{
		{ID: newID(), Name: "Check axes, belts, fans and extruder", Kind: "hours", IntervalHours: 200, LastDoneAt: stamp},
		{ID: newID(), Name: "Inspect the electrical connectors", Kind: "hours", IntervalHours: 600, LastDoneAt: stamp},
		{ID: newID(), Name: "Lubricate the linear rail", Kind: "days", IntervalDays: 90, LastDoneAt: stamp},
		{ID: newID(), Name: "Check the nozzle, replace when worn", Kind: "hours", IntervalHours: 5000, LastDoneAt: stamp, Material: "hardened steel"},
	}
}

func round1(value float64) float64 {
	return float64(int64(value*10+0.5)) / 10
}

func nowText() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func newID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)

	return hex.EncodeToString(buf)
}

func answer(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(body)
}

type stdio struct{}

func (stdio) Read(p []byte) (int, error)  { return os.Stdin.Read(p) }
func (stdio) Write(p []byte) (int, error) { return os.Stdout.Write(p) }
func (stdio) Close() error                { return os.Stdin.Close() }
