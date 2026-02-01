package policy

import (
	"testing"
)

func TestDefaultPolicy_BlockRmRfRoot(t *testing.T) {
	e := NewEngine()
	e.Policy = DefaultPolicy()

	tests := []struct {
		cmd     []string
		blocked bool
	}{
		{[]string{"rm", "-rf", "/"}, true},
		{[]string{"rm", "-fr", "/"}, true},
		{[]string{"rm", "-rf", "/tmp"}, false},
		{[]string{"rm", "-rf", "."}, false},
	}

	for _, tt := range tests {
		result, err := e.Check(tt.cmd)
		if err != nil {
			t.Fatalf("Check failed: %v", err)
		}
		if tt.blocked && result.Allowed {
			t.Errorf("expected %v to be blocked", tt.cmd)
		}
		if !tt.blocked && !result.Allowed {
			t.Errorf("expected %v to be allowed", tt.cmd)
		}
	}
}

func TestDefaultPolicy_WarnSudo(t *testing.T) {
	e := NewEngine()
	e.Policy = DefaultPolicy()

	result, err := e.Check([]string{"sudo", "apt", "update"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Allowed {
		t.Error("sudo should warn, not block")
	}
	hasWarn := false
	for _, ev := range result.Events {
		if ev.Action == "warn" && ev.Rule == "warn-sudo" {
			hasWarn = true
		}
	}
	if !hasWarn {
		t.Error("expected warn event for sudo")
	}
}

func TestDefaultPolicy_LogGitPush(t *testing.T) {
	e := NewEngine()
	e.Policy = DefaultPolicy()

	result, err := e.Check([]string{"git", "push", "origin", "main"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Allowed {
		t.Error("git push should be allowed with log")
	}
	hasLog := false
	for _, ev := range result.Events {
		if ev.Action == "log" && ev.Rule == "log-git-push" {
			hasLog = true
		}
	}
	if !hasLog {
		t.Error("expected log event for git push")
	}
}

func TestDefaultPolicy_SafeCommands(t *testing.T) {
	e := NewEngine()
	e.Policy = DefaultPolicy()

	safeCommands := [][]string{
		{"echo", "hello"},
		{"ls", "-la"},
		{"cat", "file.txt"},
		{"npm", "install"},
		{"go", "build"},
		{"git", "status"},
		{"rm", "-rf", "./node_modules"},
	}

	for _, cmd := range safeCommands {
		result, err := e.Check(cmd)
		if err != nil {
			t.Fatalf("Check failed: %v", err)
		}
		if !result.Allowed {
			t.Errorf("expected %v to be allowed", cmd)
		}
	}
}

func TestDefaultPolicyYAML(t *testing.T) {
	yaml := DefaultPolicyYAML()
	if yaml == "" {
		t.Error("DefaultPolicyYAML should not be empty")
	}
	if len(yaml) < 100 {
		t.Error("DefaultPolicyYAML seems too short")
	}
}
