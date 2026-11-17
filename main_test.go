package main

import (
	"testing"
	"time"
)

// TestDefaultsForCoreOneHoldsThePrusaSchedule keeps the seeded schedule to the
// numbers the Prusa Knowledge Base publishes for the CORE One.
func TestDefaultsForCoreOneHoldsThePrusaSchedule(t *testing.T) {
	t.Parallel()

	tasks := defaultsFor("COREONE", time.Now())
	if len(tasks) != 4 {
		t.Fatalf("the CORE One needs 4 default tasks, got %d", len(tasks))
	}

	want := map[string]float64{
		"Check axes, belts, fans and extruder": 200,
		"Inspect the electrical connectors":    600,
		"Check the nozzle, replace when worn":  250,
	}
	for _, task := range tasks {
		if task.Kind != "hours" {
			continue
		}
		if got, ok := want[task.Name]; !ok || got != task.IntervalHours {
			t.Errorf("task %q carries %v hours, want %v", task.Name, task.IntervalHours, got)
		}
	}
}

// TestAnUnknownModelStillGetsOneTask keeps the seed useful for a printer whose
// model this release does not know.
func TestAnUnknownModelStillGetsOneTask(t *testing.T) {
	t.Parallel()

	if got := defaultsFor("MK4S", time.Now()); len(got) != 1 {
		t.Fatalf("an unknown model gets one task, got %d", len(got))
	}
}

// TestTheViewMarksATaskDueSoonAtTheAlertShare holds the three states to the
// percent of the interval the operator set.
func TestTheViewMarksATaskDueSoonAtTheAlertShare(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	subject := &module{alert: 90}
	tasks := []Task{
		{ID: "a", Kind: "hours", IntervalHours: 100, LastDoneHours: 0},
		{ID: "b", Kind: "hours", IntervalHours: 100, LastDoneHours: 0},
		{ID: "c", Kind: "hours", IntervalHours: 100, LastDoneHours: 0},
	}

	views := subject.view(tasks, 50, now)
	if views[0].Status != "ok" {
		t.Errorf("50 of 100 reads %q, want ok", views[0].Status)
	}

	views = subject.view(tasks, 95, now)
	if views[0].Status != "soon" {
		t.Errorf("95 of 100 reads %q, want soon", views[0].Status)
	}

	views = subject.view(tasks, 120, now)
	if views[0].Status != "due" {
		t.Errorf("120 of 100 reads %q, want due", views[0].Status)
	}
}
