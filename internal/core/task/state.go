package task

import (
	"time"

	utiljson "github.com/user/blade-agent-runtime/internal/util/json"
)

type State struct {
	Version      int       `json:"version"`
	ActiveTaskID string    `json:"active_task_id"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func DefaultState() *State {
	return &State{
		Version:      1,
		ActiveTaskID: "",
		UpdatedAt:    time.Now().UTC(),
	}
}

func LoadState(path string) (*State, error) {
	state := DefaultState()
	if err := utiljson.ReadFile(path, state); err != nil {
		return state, err
	}
	return state, nil
}

func SaveState(path string, state *State) error {
	state.UpdatedAt = time.Now().UTC()
	return utiljson.WriteFile(path, state)
}
