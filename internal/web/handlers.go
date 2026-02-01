package web

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/blade-agent-runtime/internal/core/task"
)

type TaskResponse struct {
	*task.Task
	IsActive bool `json:"is_active"`
}

type StatusResponse struct {
	ActiveTaskID string     `json:"active_task_id"`
	ActiveTask   *task.Task `json:"active_task,omitempty"`
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := s.taskManager.List()
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	state, _ := s.taskManager.LoadState()
	response := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		response[i] = TaskResponse{
			Task:     t,
			IsActive: t.ID == state.ActiveTaskID,
		}
	}

	s.writeJSON(w, response)
}

func (s *Server) handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if taskID == "" {
		s.writeError(w, nil, http.StatusBadRequest)
		return
	}

	t, err := s.taskManager.Get(taskID)
	if err != nil {
		s.writeError(w, err, http.StatusNotFound)
		return
	}

	s.writeJSON(w, t)
}

func (s *Server) handleLedger(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/ledger/")
	if taskID == "" {
		s.writeError(w, nil, http.StatusBadRequest)
		return
	}

	entries, err := s.ledgerReader.ReadAll(taskID)
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, entries)
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/diff/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		s.writeError(w, nil, http.StatusBadRequest)
		return
	}

	taskID, stepID := parts[0], parts[1]
	// Try .patch first, then .diff (for backward compatibility if any)
	diffPath := filepath.Join(s.barDir, "tasks", taskID, "artifacts", stepID+".patch")
	if _, err := s.ledgerReader.ReadFile(diffPath); err != nil {
		diffPath = filepath.Join(s.barDir, "tasks", taskID, "artifacts", stepID+".diff")
	}

	data, err := s.ledgerReader.ReadFile(diffPath)
	if err != nil {
		s.writeError(w, err, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	state, err := s.taskManager.LoadState()
	if err != nil {
		s.writeError(w, err, http.StatusInternalServerError)
		return
	}

	response := StatusResponse{ActiveTaskID: state.ActiveTaskID}
	if state.ActiveTaskID != "" {
		response.ActiveTask, _ = s.taskManager.Get(state.ActiveTaskID)
	}

	s.writeJSON(w, response)
}

func (s *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	msg := http.StatusText(code)
	if err != nil {
		msg = err.Error()
	}
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
