package web

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/user/blade-agent-runtime/internal/core/diff"
	"github.com/user/blade-agent-runtime/internal/core/ledger"
	"github.com/user/blade-agent-runtime/internal/core/task"
)

type Server struct {
	addr         string
	taskManager  *task.Manager
	ledgerReader *ledger.Reader
	barDir       string
	wsHub        *WebSocketHub
	httpServer   *http.Server
}

func NewServer(addr string, taskManager *task.Manager, barDir string) *Server {
	return &Server{
		addr:         addr,
		taskManager:  taskManager,
		ledgerReader: ledger.NewReader(filepath.Join(barDir, "tasks")),
		barDir:       barDir,
		wsHub:        NewWebSocketHub(),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         s.addr,
		Handler:      s.middleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go s.wsHub.Run()

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskDetail)
	mux.HandleFunc("/api/ledger/", s.handleLedger)
	mux.HandleFunc("/api/diff/", s.handleDiff)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/ws", s.wsHub.HandleWebSocket)
	mux.HandleFunc("/", s.handleStatic)
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Broadcast(msgType string, data interface{}) {
	s.wsHub.Broadcast(msgType, data)
}

func (s *Server) BroadcastLiveDiff(taskID string, result *diff.Result) {
	s.wsHub.Broadcast("live_diff", map[string]interface{}{
		"task_id":   taskID,
		"files":     result.Files,
		"additions": result.Additions,
		"deletions": result.Deletions,
		"file_list": result.FileList,
		"patch":     string(result.Patch),
	})
}
