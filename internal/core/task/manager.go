package task

import (
	"os"
	"path/filepath"
	"time"

	barerrors "github.com/user/blade-agent-runtime/internal/util/errors"
	utiljson "github.com/user/blade-agent-runtime/internal/util/json"
)

type Manager struct {
	RepoRoot  string
	BarDir    string
	TasksDir  string
	StatePath string
}

func NewManager(repoRoot string, barDir string) *Manager {
	return &Manager{
		RepoRoot:  repoRoot,
		BarDir:    barDir,
		TasksDir:  filepath.Join(barDir, "tasks"),
		StatePath: filepath.Join(barDir, "state.json"),
	}
}

func (m *Manager) Create(id string, name string, baseRef string, branch string, workspacePath string) (*Task, error) {
	now := time.Now().UTC()
	task := &Task{
		ID:            id,
		Name:          name,
		RepoRoot:      m.RepoRoot,
		BaseRef:       baseRef,
		Branch:        branch,
		WorkspacePath: workspacePath,
		Status:        TaskStatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	taskDir := filepath.Join(m.TasksDir, id)
	if err := os.MkdirAll(filepath.Join(taskDir, "artifacts"), 0o755); err != nil {
		return nil, err
	}
	if err := utiljson.WriteFile(filepath.Join(taskDir, "task.json"), task); err != nil {
		return nil, err
	}
	ledgerPath := filepath.Join(taskDir, "ledger.jsonl")
	if _, err := os.Stat(ledgerPath); os.IsNotExist(err) {
		if err := os.WriteFile(ledgerPath, []byte{}, 0o644); err != nil {
			return nil, err
		}
	}
	return task, nil
}

func (m *Manager) Get(taskID string) (*Task, error) {
	taskPath := filepath.Join(m.TasksDir, taskID, "task.json")
	task := &Task{}
	if err := utiljson.ReadFile(taskPath, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (m *Manager) List() ([]*Task, error) {
	entries, err := os.ReadDir(m.TasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Task{}, nil
		}
		return nil, err
	}
	tasks := []*Task{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		task, err := m.Get(entry.Name())
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (m *Manager) LoadState() (*State, error) {
	if _, err := os.Stat(m.StatePath); err != nil {
		if os.IsNotExist(err) {
			return DefaultState(), nil
		}
		return nil, err
	}
	return LoadState(m.StatePath)
}

func (m *Manager) SaveState(state *State) error {
	return SaveState(m.StatePath, state)
}

func (m *Manager) SetActive(taskID string) error {
	state, err := m.LoadState()
	if err != nil {
		return err
	}
	state.ActiveTaskID = taskID
	return m.SaveState(state)
}

func (m *Manager) GetActive() (*Task, error) {
	state, err := m.LoadState()
	if err != nil {
		return nil, err
	}
	if state.ActiveTaskID == "" {
		return nil, barerrors.NoActiveTask()
	}
	return m.Get(state.ActiveTaskID)
}

func (m *Manager) Close(task *Task) error {
	now := time.Now().UTC()
	task.Status = TaskStatusClosed
	task.ClosedAt = &now
	task.UpdatedAt = now
	taskPath := filepath.Join(m.TasksDir, task.ID, "task.json")
	return utiljson.WriteFile(taskPath, task)
}

func (m *Manager) Delete(taskID string) error {
	taskDir := filepath.Join(m.TasksDir, taskID)
	return os.RemoveAll(taskDir)
}

func (m *Manager) Update(task *Task) error {
	task.UpdatedAt = time.Now().UTC()
	taskPath := filepath.Join(m.TasksDir, task.ID, "task.json")
	return utiljson.WriteFile(taskPath, task)
}

func (m *Manager) ResolveByName(name string) (*Task, error) {
	tasks, err := m.List()
	if err != nil {
		return nil, err
	}
	for _, t := range tasks {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, barerrors.TaskNotFound(name)
}
