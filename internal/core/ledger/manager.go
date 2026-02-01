package ledger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

type Manager struct {
	TaskDir string
}

func NewManager(taskDir string) *Manager {
	return &Manager{TaskDir: taskDir}
}

func (m *Manager) LedgerPath() string {
	return filepath.Join(m.TaskDir, "ledger.jsonl")
}

func (m *Manager) Append(step *Step) error {
	f, err := os.OpenFile(m.LedgerPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(step)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

func (m *Manager) List() ([]*Step, error) {
	f, err := os.Open(m.LedgerPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []*Step{}, nil
		}
		return nil, err
	}
	defer f.Close()
	steps := []*Step{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var step Step
		if err := json.Unmarshal(line, &step); err != nil {
			return nil, err
		}
		steps = append(steps, &step)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return steps, nil
}

func (m *Manager) GetLast() (*Step, error) {
	steps, err := m.List()
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		return nil, nil
	}
	return steps[len(steps)-1], nil
}

func (m *Manager) GetByID(stepID string) (*Step, error) {
	steps, err := m.List()
	if err != nil {
		return nil, err
	}
	for _, s := range steps {
		if s.StepID == stepID {
			return s, nil
		}
	}
	return nil, nil
}

func (m *Manager) NextStepID() (string, error) {
	last, err := m.GetLast()
	if err != nil {
		return "", err
	}
	if last == nil || last.StepID == "" {
		return "0001", nil
	}
	num, err := strconv.Atoi(last.StepID)
	if err != nil {
		return "", err
	}
	return formatStepID(num + 1), nil
}

func formatStepID(n int) string {
	return strconv.FormatInt(int64(10000+n), 10)[1:]
}
