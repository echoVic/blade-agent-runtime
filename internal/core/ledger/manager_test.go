package ledger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_AppendAndList(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "artifacts"), 0755)

	m := NewManager(tmpDir)

	step := &Step{
		StepID:    "0001",
		Kind:      "run",
		Cmd:       []string{"echo", "hello"},
		Cwd:       ".",
		StartedAt: time.Now().UTC(),
		EndedAt:   time.Now().UTC(),
		ExitCode:  intPtr(0),
		DiffStat: &DiffStat{
			Files:     1,
			Additions: 10,
			Deletions: 2,
		},
	}

	if err := m.Append(step); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	steps, err := m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(steps))
	}

	if steps[0].StepID != "0001" {
		t.Errorf("expected StepID '0001', got '%s'", steps[0].StepID)
	}
	if steps[0].Kind != "run" {
		t.Errorf("expected Kind 'run', got '%s'", steps[0].Kind)
	}
}

func TestManager_GetLast(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "artifacts"), 0755)

	m := NewManager(tmpDir)

	step1 := &Step{StepID: "0001", Kind: "run", StartedAt: time.Now(), EndedAt: time.Now()}
	step2 := &Step{StepID: "0002", Kind: "run", StartedAt: time.Now(), EndedAt: time.Now()}
	step3 := &Step{StepID: "0003", Kind: "apply", StartedAt: time.Now(), EndedAt: time.Now()}

	_ = m.Append(step1)
	_ = m.Append(step2)
	_ = m.Append(step3)

	last, err := m.GetLast()
	if err != nil {
		t.Fatalf("GetLast failed: %v", err)
	}

	if last.StepID != "0003" {
		t.Errorf("expected last StepID '0003', got '%s'", last.StepID)
	}
}

func TestManager_GetByID(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "artifacts"), 0755)

	m := NewManager(tmpDir)

	step1 := &Step{StepID: "0001", Kind: "run", StartedAt: time.Now(), EndedAt: time.Now()}
	step2 := &Step{StepID: "0002", Kind: "rollback", StartedAt: time.Now(), EndedAt: time.Now()}

	_ = m.Append(step1)
	_ = m.Append(step2)

	got, err := m.GetByID("0001")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Kind != "run" {
		t.Errorf("expected Kind 'run', got '%s'", got.Kind)
	}

	got2, err := m.GetByID("0002")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got2.Kind != "rollback" {
		t.Errorf("expected Kind 'rollback', got '%s'", got2.Kind)
	}

	got3, err := m.GetByID("9999")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got3 != nil {
		t.Error("expected nil for nonexistent step")
	}
}

func TestManager_NextStepID(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "artifacts"), 0755)

	m := NewManager(tmpDir)

	id1, err := m.NextStepID()
	if err != nil {
		t.Fatalf("NextStepID failed: %v", err)
	}
	if id1 != "0001" {
		t.Errorf("expected '0001', got '%s'", id1)
	}

	_ = m.Append(&Step{StepID: "0001", Kind: "run", StartedAt: time.Now(), EndedAt: time.Now()})

	id2, err := m.NextStepID()
	if err != nil {
		t.Fatalf("NextStepID failed: %v", err)
	}
	if id2 != "0002" {
		t.Errorf("expected '0002', got '%s'", id2)
	}
}

func TestManager_EmptyLedger(t *testing.T) {
	tmpDir := t.TempDir()

	m := NewManager(tmpDir)

	steps, err := m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(steps))
	}

	last, err := m.GetLast()
	if err != nil {
		t.Fatalf("GetLast failed: %v", err)
	}
	if last != nil {
		t.Error("expected nil for empty ledger")
	}
}

func intPtr(i int) *int {
	return &i
}
