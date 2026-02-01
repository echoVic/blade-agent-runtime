package policy

import (
	"testing"
)

func TestEngine_Check_Block(t *testing.T) {
	e := NewEngine()
	e.Policy = &Policy{
		Version: 1,
		Rules: []Rule{
			{Name: "no-rm-rf-root", Pattern: `rm\s+-rf\s+/`, Action: "block", Reason: "Dangerous"},
			{Name: "no-rm-rf-home", Pattern: `rm\s+-rf\s+~`, Action: "block", Reason: "Dangerous"},
		},
	}

	result, err := e.Check([]string{"rm", "-rf", "/"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if result.Allowed {
		t.Error("expected command to be blocked")
	}
	if len(result.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(result.Events))
	}
	if result.Events[0].Rule != "no-rm-rf-root" {
		t.Errorf("expected rule 'no-rm-rf-root', got '%s'", result.Events[0].Rule)
	}
}

func TestEngine_Check_Allow(t *testing.T) {
	e := NewEngine()
	e.Policy = &Policy{
		Version: 1,
		Rules: []Rule{
			{Name: "no-rm-rf-root", Pattern: `rm\s+-rf\s+/`, Action: "block", Reason: "Dangerous"},
		},
	}

	result, err := e.Check([]string{"echo", "hello"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Allowed {
		t.Error("expected command to be allowed")
	}
	if len(result.Events) != 0 {
		t.Errorf("expected 0 events, got %d", len(result.Events))
	}
}

func TestEngine_Check_Warn(t *testing.T) {
	e := NewEngine()
	e.Policy = &Policy{
		Version: 1,
		Rules: []Rule{
			{Name: "warn-sudo", Pattern: `^sudo\s+`, Action: "warn", Reason: "Elevated privileges"},
		},
	}

	result, err := e.Check([]string{"sudo", "apt", "update"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Allowed {
		t.Error("warn should not block command")
	}
	if len(result.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(result.Events))
	}
	if result.Events[0].Action != "warn" {
		t.Errorf("expected action 'warn', got '%s'", result.Events[0].Action)
	}
}

func TestEngine_Check_MultipleRules(t *testing.T) {
	e := NewEngine()
	e.Policy = &Policy{
		Version: 1,
		Rules: []Rule{
			{Name: "warn-sudo", Pattern: `^sudo`, Action: "warn", Reason: "Elevated"},
			{Name: "block-rm-rf", Pattern: `rm\s+-rf`, Action: "block", Reason: "Dangerous"},
		},
	}

	result, err := e.Check([]string{"sudo", "rm", "-rf", "/tmp/test"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if result.Allowed {
		t.Error("expected command to be blocked")
	}
	if len(result.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(result.Events))
	}
}

func TestEngine_Check_NilPolicy(t *testing.T) {
	e := NewEngine()

	result, err := e.Check([]string{"rm", "-rf", "/"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Allowed {
		t.Error("nil policy should allow all commands")
	}
}

func TestEngine_Check_EmptyRules(t *testing.T) {
	e := NewEngine()
	e.Policy = &Policy{
		Version: 1,
		Rules:   []Rule{},
	}

	result, err := e.Check([]string{"any", "command"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Allowed {
		t.Error("empty rules should allow all commands")
	}
}

func TestEngine_Check_InvalidRegex(t *testing.T) {
	e := NewEngine()
	e.Policy = &Policy{
		Version: 1,
		Rules: []Rule{
			{Name: "invalid", Pattern: `[invalid`, Action: "block", Reason: "Test"},
		},
	}

	_, err := e.Check([]string{"test"})
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestEngine_Check_LogAction(t *testing.T) {
	e := NewEngine()
	e.Policy = &Policy{
		Version: 1,
		Rules: []Rule{
			{Name: "log-git", Pattern: `^git`, Action: "log", Reason: "Logging git commands"},
		},
	}

	result, err := e.Check([]string{"git", "status"})
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if !result.Allowed {
		t.Error("log action should not block command")
	}
	if len(result.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(result.Events))
	}
	if result.Events[0].Action != "log" {
		t.Errorf("expected action 'log', got '%s'", result.Events[0].Action)
	}
}
