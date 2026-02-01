package task

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_CreateAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	barDir := filepath.Join(tmpDir, ".bar")
	os.MkdirAll(filepath.Join(barDir, "tasks"), 0755)

	m := NewManager(tmpDir, barDir)

	task, err := m.Create("test123", "fix-bug", "main", "bar/fix-bug-test123", "/workspace/test123")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if task.ID != "test123" {
		t.Errorf("expected ID 'test123', got '%s'", task.ID)
	}
	if task.Name != "fix-bug" {
		t.Errorf("expected Name 'fix-bug', got '%s'", task.Name)
	}
	if task.Status != "active" {
		t.Errorf("expected Status 'active', got '%s'", task.Status)
	}

	got, err := m.Get("test123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("Get returned wrong task")
	}
}

func TestManager_List(t *testing.T) {
	tmpDir := t.TempDir()
	barDir := filepath.Join(tmpDir, ".bar")
	os.MkdirAll(filepath.Join(barDir, "tasks"), 0755)

	m := NewManager(tmpDir, barDir)

	_, _ = m.Create("task1", "task-one", "main", "bar/task-one", "/ws/task1")
	_, _ = m.Create("task2", "task-two", "main", "bar/task-two", "/ws/task2")

	tasks, err := m.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestManager_Close(t *testing.T) {
	tmpDir := t.TempDir()
	barDir := filepath.Join(tmpDir, ".bar")
	os.MkdirAll(filepath.Join(barDir, "tasks"), 0755)

	m := NewManager(tmpDir, barDir)

	task, _ := m.Create("closetest", "to-close", "main", "bar/to-close", "/ws/closetest")

	if err := m.Close(task); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	got, _ := m.Get("closetest")
	if got.Status != "closed" {
		t.Errorf("expected Status 'closed', got '%s'", got.Status)
	}
	if got.ClosedAt == nil {
		t.Error("ClosedAt should not be nil after close")
	}
}

func TestManager_State(t *testing.T) {
	tmpDir := t.TempDir()
	barDir := filepath.Join(tmpDir, ".bar")
	os.MkdirAll(barDir, 0755)

	m := NewManager(tmpDir, barDir)

	state := &State{
		Version:      1,
		ActiveTaskID: "abc123",
		UpdatedAt:    time.Now().UTC(),
	}

	if err := m.SaveState(state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	loaded, err := m.LoadState()
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.ActiveTaskID != "abc123" {
		t.Errorf("expected ActiveTaskID 'abc123', got '%s'", loaded.ActiveTaskID)
	}
}

func TestManager_SetActive(t *testing.T) {
	tmpDir := t.TempDir()
	barDir := filepath.Join(tmpDir, ".bar")
	os.MkdirAll(filepath.Join(barDir, "tasks"), 0755)

	m := NewManager(tmpDir, barDir)

	_, _ = m.Create("active1", "task-active", "main", "bar/task-active", "/ws/active1")

	if err := m.SetActive("active1"); err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	state, _ := m.LoadState()
	if state.ActiveTaskID != "active1" {
		t.Errorf("expected ActiveTaskID 'active1', got '%s'", state.ActiveTaskID)
	}
}

func TestManager_ResolveByName(t *testing.T) {
	tmpDir := t.TempDir()
	barDir := filepath.Join(tmpDir, ".bar")
	os.MkdirAll(filepath.Join(barDir, "tasks"), 0755)

	m := NewManager(tmpDir, barDir)

	_, _ = m.Create("resolve1", "my-task", "main", "bar/my-task", "/ws/resolve1")

	task, err := m.ResolveByName("my-task")
	if err != nil {
		t.Fatalf("ResolveByName failed: %v", err)
	}
	if task.ID != "resolve1" {
		t.Errorf("expected ID 'resolve1', got '%s'", task.ID)
	}

	_, err = m.ResolveByName("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent task name")
	}
}

func TestManager_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	barDir := filepath.Join(tmpDir, ".bar")
	os.MkdirAll(filepath.Join(barDir, "tasks"), 0755)

	m := NewManager(tmpDir, barDir)

	_, _ = m.Create("todelete", "delete-me", "main", "bar/delete-me", "/ws/todelete")

	if err := m.Delete("todelete"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := m.Get("todelete")
	if err == nil {
		t.Error("expected error after delete")
	}
}
