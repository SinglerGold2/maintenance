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
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"github.com/prusa3d/prusa-connect/sdk/busapi"
	"github.com/prusa3d/prusa-connect/sdk/deviceapi"
	"github.com/prusa3d/prusa-connect/sdk/extension"
	"github.com/prusa3d/prusa-connect/sdk/moduledb"
	"github.com/prusa3d/prusa-connect/sdk/routesapi"
	"github.com/prusa3d/prusa-connect/sdk/storageapi"
	"github.com/prusa3d/prusa-connect/sdk/wire"
	"github.com/prusa3d/prusa-connect/sdk/wirerpc"

	"github.com/singlergold2/maintenance/internal/gcode"
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
	Builtin       bool    `json:"builtin,omitempty"`
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
	PrinterHours *float64    `json:"printer_hours,omitempty"`
	Tools        []ToolHours `json:"tools,omitempty"`
	Tasks        []TaskView  `json:"tasks"`
}

// ToolHours is one tool of a printer and the hours it ran.
type ToolHours struct {
	Tool  int     `json:"tool"`
	Hours float64 `json:"hours"`
}

// module is the state the routes read.
type module struct {
	mu       sync.Mutex
	db       *sql.DB
	postgres bool
	devices  *deviceapi.Client
	bus      *busapi.Client
	storage  *storageapi.Client
	alert    int

	notified map[string]string

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
	state.bus = busapi.NewClient(conn)
	state.storage = storageapi.NewClient(conn)
	state.mu.Unlock()

	fmt.Fprintf(os.Stderr, "%s %s is serving on stdio as pid %d\n", name, release, os.Getpid())

	if err = conn.Serve(ctx); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

//nolint:gochecknoglobals // the one piece of state this module has.
var state = &module{
	alert:    defaultAlertPercent,
	cache:    map[string]hoursEntry{},
	notified: map[string]string{},
}

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
	mux.HandleFunc("POST /devices/tools/reset", handleResetTool)
	mux.HandleFunc("GET /template", handleTemplate)
	mux.HandleFunc("POST /template", handleSaveTemplate)

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

	devices, err := state.listDevices(r.Context(), routesapi.PrincipalOf(r))
	if err != nil {
		refuse(w, r, HistoryFailed.New("the fleet could not be read").With(faultString("reason", err.Error())))
		return
	}

	views := make([]DeviceView, 0, len(devices))

	for _, device := range devices {
		hours, prints, hoursErr := state.hoursOf(r.Context(), device.ID, routesapi.PrincipalOf(r))
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
				schedule, tplErr := st.Template(r.Context())
				if tplErr != nil {
					refuse(w, r, StoreFailed.New("the schedule could not be read").With(faultString("reason", tplErr.Error())))
					return
				}

				tasks = tasksFrom(schedule, device.ID, time.Now().UTC())
				if seedErr = st.Insert(r.Context(), tasks); seedErr != nil {
					refuse(w, r, StoreFailed.New("the default tasks could not be written").With(faultString("reason", seedErr.Error())))
					return
				}
			}
		}

		printerTotal, offset, known, hoursErr := st.PrinterCounts(r.Context(), device.ID)
		if hoursErr != nil {
			refuse(w, r, StoreFailed.New("the printer's own hours could not be read").With(faultString("reason", hoursErr.Error())))
			return
		}

		counter := hours

		var typed *float64
		if known {
			if offset < 0 {
				offset = printerTotal - hours
				if offsetErr := st.SetPrinterOffset(r.Context(), device.ID, offset); offsetErr != nil {
					refuse(w, r, StoreFailed.New("the printer's lead could not be written").With(faultString("reason", offsetErr.Error())))
					return
				}
			}

			// The printer keeps the lead it had when the operator typed its
			// value in, so both counters move with every print.
			projected := hours + offset
			typed = &projected
			counter = projected
		}

		if attributeErr := state.attributeTools(r.Context(), st, device.ID, routesapi.PrincipalOf(r)); attributeErr != nil {
			fmt.Fprintf(os.Stderr, "%s: the tools of %s: %v\n", name, device.ID, attributeErr)
		}

		toolMap, toolsErr := st.ToolHours(r.Context(), device.ID)
		if toolsErr != nil {
			refuse(w, r, StoreFailed.New("the tool hours could not be read").With(faultString("reason", toolsErr.Error())))
			return
		}

		// One tool takes every hour the printer ran: nothing is split, and the
		// printer's own counter is the honest figure, less whatever a reset
		// already put behind it.
		if len(toolMap) <= 1 {
			reset, hasReset, resetErr := st.ToolReset(r.Context(), device.ID, 0)
			if resetErr != nil {
				refuse(w, r, StoreFailed.New("the tool reset could not be read").With(faultString("reason", resetErr.Error())))
				return
			}

			single := counter
			if hasReset {
				single = counter - reset
			}

			if single < 0 {
				single = 0
			}

			toolMap = map[int]float64{0: single}
		}

		tools := make([]ToolHours, 0, len(toolMap))
		for tool, toolHours := range toolMap {
			tools = append(tools, ToolHours{Tool: tool, Hours: round1(toolHours)})
		}

		sort.Slice(tools, func(i, j int) bool { return tools[i].Tool < tools[j].Tool })

		deviceTasks := state.view(tasks, counter, time.Now())
		state.announce(r.Context(), device.ID, device.Name, deviceTasks)

		views = append(views, DeviceView{
			ID:           device.ID,
			Name:         device.Name,
			Model:        device.Model,
			Hours:        round1(hours),
			Prints:       prints,
			PrinterHours: typed,
			Tools:        tools,
			Tasks:        deviceTasks,
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

	hours, _, hoursErr := state.hoursOf(r.Context(), task.DeviceID, routesapi.PrincipalOf(r))
	if hoursErr != nil {
		refuse(w, r, HistoryFailed.New("the print history could not be read").With(faultString("reason", hoursErr.Error())))
		return
	}

	_, offset, known, countsErr := st.PrinterCounts(r.Context(), task.DeviceID)
	if countsErr != nil {
		refuse(w, r, StoreFailed.New("the printer's own hours could not be read").With(faultString("reason", countsErr.Error())))
		return
	}

	counter := hours
	if known && offset >= 0 {
		counter = hours + offset
	}

	task.LastDoneHours = counter
	task.LastDoneAt = time.Now().UTC().Format(time.RFC3339)

	if saveErr := st.Upsert(r.Context(), task); saveErr != nil {
		refuse(w, r, StoreFailed.New("the task could not be written").With(faultString("reason", saveErr.Error())))
		return
	}

	answer(w, http.StatusOK, task)
}

func handleTemplate(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	tasks, listErr := st.Template(r.Context())
	if listErr != nil {
		refuse(w, r, StoreFailed.New("the schedule could not be read").With(faultString("reason", listErr.Error())))
		return
	}

	answer(w, http.StatusOK, map[string]any{"tasks": tasks, "alert_percent": state.alertPercent()})
}

func handleSaveTemplate(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	var sent TemplateTask
	if decodeErr := json.NewDecoder(r.Body).Decode(&sent); decodeErr != nil || sent.ID == "" {
		refuse(w, r, InvalidInput.New("the request names no schedule line"))
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

	if saveErr := st.UpdateTemplate(r.Context(), sent); saveErr != nil {
		refuse(w, r, StoreFailed.New("the schedule could not be written").With(faultString("reason", saveErr.Error())))
		return
	}

	answer(w, http.StatusOK, sent)
}

func handleResetTool(w http.ResponseWriter, r *http.Request) {
	st, err := state.store()
	if err != nil {
		refuse(w, r, NoDatabase.New("this instance has no database"))
		return
	}

	var sent struct {
		DeviceID string `json:"device_id"`
		Tool     int    `json:"tool"`
	}

	if decodeErr := json.NewDecoder(r.Body).Decode(&sent); decodeErr != nil || sent.DeviceID == "" {
		refuse(w, r, InvalidInput.New("the request names no printer"))
		return
	}

	hours, _, hoursErr := state.hoursOf(r.Context(), sent.DeviceID, routesapi.PrincipalOf(r))
	if hoursErr != nil {
		refuse(w, r, HistoryFailed.New("the print history could not be read").With(faultString("reason", hoursErr.Error())))
		return
	}

	_, offset, known, countsErr := st.PrinterCounts(r.Context(), sent.DeviceID)
	if countsErr != nil {
		refuse(w, r, StoreFailed.New("the printer's own hours could not be read").With(faultString("reason", countsErr.Error())))
		return
	}

	counter := hours
	if known && offset >= 0 {
		counter = hours + offset
	}

	toolMap, toolErr := st.ToolHours(r.Context(), sent.DeviceID)
	if toolErr != nil {
		refuse(w, r, StoreFailed.New("the tool hours could not be read").With(faultString("reason", toolErr.Error())))
		return
	}

	at := counter
	if len(toolMap) > 1 {
		at = toolMap[sent.Tool]
	}

	if resetErr := st.ResetTool(r.Context(), sent.DeviceID, sent.Tool, at); resetErr != nil {
		refuse(w, r, StoreFailed.New("the tool could not be reset").With(faultString("reason", resetErr.Error())))
		return
	}

	answer(w, http.StatusOK, map[string]any{"device_id": sent.DeviceID, "tool": sent.Tool, "hours": 0})
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

	tracked, _, trackedErr := state.hoursOf(r.Context(), sent.DeviceID, routesapi.PrincipalOf(r))
	if trackedErr != nil {
		refuse(w, r, HistoryFailed.New("the print history could not be read").With(faultString("reason", trackedErr.Error())))
		return
	}

	if storeErr := st.SetPrinterHours(r.Context(), sent.DeviceID, sent.Hours, tracked); storeErr != nil {
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

	task, foundErr := st.Task(r.Context(), id)
	if foundErr != nil {
		refuse(w, r, NoTask.New("no task carries that id").With(faultString("id", id)))
		return
	}

	if task.Builtin {
		refuse(w, r, BuiltinTask.New("this task ships with the module and cannot be removed"))
		return
	}

	if deleteErr := st.Delete(r.Context(), id); deleteErr != nil {
		refuse(w, r, StoreFailed.New("the task could not be removed").With(faultString("reason", deleteErr.Error())))
		return
	}

	answer(w, http.StatusOK, map[string]any{"deleted": id})
}

const (
	// toolFileLimit keeps a page responsive: a few files per load are read, the
	// rest on the loads after it.
	toolFileLimit = 3
	maxToolFile   = 512 << 20
)

// attributeTools splits the length of the finished prints this module has not
// read yet across the tools the G-code names. It reads at most a few files per
// call, so a page stays quick and the older prints follow on later loads.
func (m *module) attributeTools(ctx context.Context, st *store, deviceID, principal string) error {
	m.mu.Lock()
	devices := m.devices
	files := m.storage
	m.mu.Unlock()

	if devices == nil || files == nil {
		return nil
	}

	read := 0
	cursor := ""

	for page := 0; page < 5 && read < toolFileLimit; page++ {
		history, err := devices.JobHistory(ctx, deviceapi.JobHistoryParams{Principal: principal, ID: deviceID, Limit: 50, Cursor: cursor})
		if err != nil {
			return err
		}

		for _, job := range history.Items {
			if job.Ended == nil || job.Path == "" {
				continue
			}

			seen, seenErr := st.JobSeen(ctx, deviceID, job.ID)
			if seenErr != nil {
				return seenErr
			}

			if seen {
				continue
			}

			if read >= toolFileLimit {
				return nil
			}

			read++

			if splitErr := m.splitJob(ctx, st, files, deviceID, job); splitErr != nil {
				fmt.Fprintf(os.Stderr, "%s: print %s: %v\n", name, job.ID, splitErr)
			}
		}

		if history.Cursor == "" || history.Cursor == cursor {
			break
		}

		cursor = history.Cursor
	}

	return nil
}

// splitJob reads one print's G-code and adds each tool's share of the print to
// its counter. A file this module cannot read is marked seen all the same, so
// the same broken file is not opened on every load.
func (m *module) splitJob(ctx context.Context, st *store, files *storageapi.Client, deviceID string, job deviceapi.Job) error {
	file, err := files.Open(ctx, storageapi.Target{Storage: "local", Path: job.Path})
	if err != nil && errors.Is(err, storageapi.ErrNoFile) {
		// The print is gone from the storage. Nothing to read, now or later.
		_ = st.MarkJobSeen(ctx, deviceID, job.ID)

		return nil
	}

	if err != nil {
		// The printer's own path carries its folder ("usb/…"), the server's
		// storage keeps the file under its name alone. Try that before giving up.
		base := path.Base(job.Path)
		if base == job.Path {
			return err
		}

		file, err = files.Open(ctx, storageapi.Target{Storage: "local", Path: base})
		if err != nil {
			return err
		}
	}

	if file.Size() > maxToolFile {
		return fmt.Errorf("the file is %d bytes, larger than the %d this module reads", file.Size(), maxToolFile)
	}

	fmt.Fprintf(os.Stderr, "%s: print %s: reading %s, %d bytes\n", name, job.ID, job.Path, file.Size())

	text, err := gcode.ReadAll(file, file.Size())
	if err != nil {
		_ = st.MarkJobSeen(ctx, deviceID, job.ID)

		return err
	}

	usage := gcode.Parse(text)
	fmt.Fprintf(os.Stderr, "%s: print %s: %d lines, %d tools\n", name, job.ID, usage.Lines, len(usage.Extrusion))

	if usage.Lines == 0 {
		return nil
	}

	var total float64
	for _, millimetres := range usage.Extrusion {
		total += millimetres
	}

	if total <= 0 || job.Ended == nil {
		return nil
	}

	length := job.Ended.Sub(job.Started).Hours()
	fmt.Fprintf(os.Stderr, "%s: print %s: %d mm, %.4f hours\n", name, job.ID, int(total), length)

	if length <= 0 {
		return nil
	}

	for tool, millimetres := range usage.Extrusion {
		share := length * millimetres / total
		if err := st.AddToolHours(ctx, deviceID, tool, share); err != nil {
			return err
		}
	}

	return st.MarkJobSeen(ctx, deviceID, job.ID)
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

func (m *module) listDevices(ctx context.Context, principal string) ([]deviceapi.Device, error) {
	m.mu.Lock()
	client := m.devices
	m.mu.Unlock()

	if client == nil {
		return nil, errors.New("no device client")
	}

	page, err := client.List(ctx, deviceapi.ListParams{Principal: principal, Limit: 100})
	if err != nil {
		return nil, err
	}

	return page.Items, nil
}

// hoursOf sums the length of every ended print of one device. The result is
// held for a few minutes, because a page load asks for it once per device.
func (m *module) hoursOf(ctx context.Context, deviceID, principal string) (float64, int, error) {
	m.cacheMu.Lock()
	key := principal + "/" + deviceID
	cached, found := m.cache[key]
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
		history, err := client.JobHistory(ctx, deviceapi.JobHistoryParams{Principal: principal, ID: deviceID, Limit: 200, Cursor: cursor})
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
	m.cache[key] = hoursEntry{hours: total, prints: prints, calculated: time.Now()}
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

// tasksFrom turns the schedule template into the tasks of one printer.
func tasksFrom(schedule []TemplateTask, deviceID string, now time.Time) []Task {
	stamp := now.Format(time.RFC3339)
	out := make([]Task, 0, len(schedule))

	for _, line := range schedule {
		out = append(out, Task{
			ID:            newID(),
			DeviceID:      deviceID,
			Name:          line.Name,
			Kind:          line.Kind,
			IntervalHours: line.IntervalHours,
			IntervalDays:  line.IntervalDays,
			Material:      line.Material,
			LastDoneAt:    stamp,
			Builtin:       true,
		})
	}

	return out
}

// defaultsFor is the maintenance schedule Prusa publishes for a model. The
// numbers are from the Prusa Knowledge Base article on regular printer
// maintenance.
func defaultsFor(model, deviceID string, now time.Time) []Task {
	upper := strings.ToUpper(strings.ReplaceAll(model, " ", ""))
	stamp := now.Format(time.RFC3339)

	if !strings.Contains(upper, "COREONE") {
		return []Task{
			{ID: newID(), DeviceID: deviceID, Builtin: true, Name: "Inspect the machine", Kind: "hours", IntervalHours: 200, LastDoneAt: stamp},
		}
	}

	return []Task{
		{ID: newID(), DeviceID: deviceID, Builtin: true, Name: "Check axes, belts, fans and extruder", Kind: "hours", IntervalHours: 200, LastDoneAt: stamp},
		{ID: newID(), DeviceID: deviceID, Builtin: true, Name: "Inspect the electrical connectors", Kind: "hours", IntervalHours: 600, LastDoneAt: stamp},
		{ID: newID(), DeviceID: deviceID, Builtin: true, Name: "Lubricate the linear rail", Kind: "days", IntervalDays: 90, LastDoneAt: stamp},
		{ID: newID(), DeviceID: deviceID, Builtin: true, Name: "Check the nozzle, replace when worn", Kind: "hours", IntervalHours: 5000, LastDoneAt: stamp, Material: "hardened steel"},
	}
}

// announce sends one event when a task reaches due soon or due. The state is
// remembered, so a page that reloads ten times announces once.
func (m *module) announce(ctx context.Context, deviceID, deviceName string, tasks []TaskView) {
	m.mu.Lock()
	bus := m.bus
	m.mu.Unlock()

	if bus == nil {
		return
	}

	for _, task := range tasks {
		key := deviceID + "/" + task.ID

		last := m.notified[key]
		if last == task.Status {
			continue
		}

		m.notified[key] = task.Status
		if task.Status != "due" && task.Status != "soon" {
			continue
		}

		_ = bus.Emit(ctx, "maintenance", key, map[string]any{
			"device_id":   deviceID,
			"device_name": deviceName,
			"task_id":     task.ID,
			"task_name":   task.Name,
			"status":      task.Status,
			"percent":     task.Percent,
			"material":    task.Material,
		})
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
