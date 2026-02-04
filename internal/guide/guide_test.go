package guide

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrompt_Select(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		input    string
		wantIdx  int
		wantErr  bool
	}{
		{
			name: "select first option",
			options: []Option{
				{Label: "Option A", Value: "a"},
				{Label: "Option B", Value: "b"},
			},
			input:   "1\n",
			wantIdx: 0,
		},
		{
			name: "select second option",
			options: []Option{
				{Label: "Option A", Value: "a"},
				{Label: "Option B", Value: "b"},
			},
			input:   "2\n",
			wantIdx: 1,
		},
		{
			name: "invalid input then valid",
			options: []Option{
				{Label: "Option A", Value: "a"},
				{Label: "Option B", Value: "b"},
			},
			input:   "invalid\n1\n",
			wantIdx: 0,
		},
		{
			name: "out of range then valid",
			options: []Option{
				{Label: "Option A", Value: "a"},
				{Label: "Option B", Value: "b"},
			},
			input:   "5\n2\n",
			wantIdx: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			p := NewPrompt(in, out)

			idx, err := p.Select("Choose:", tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if idx != tt.wantIdx {
				t.Errorf("Select() = %v, want %v", idx, tt.wantIdx)
			}
		})
	}
}

func TestPrompt_Confirm(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{name: "yes lowercase", input: "y\n", want: true},
		{name: "yes uppercase", input: "Y\n", want: true},
		{name: "yes full", input: "yes\n", want: true},
		{name: "no lowercase", input: "n\n", want: false},
		{name: "no uppercase", input: "N\n", want: false},
		{name: "no full", input: "no\n", want: false},
		{name: "empty defaults to no", input: "\n", want: false},
		{name: "invalid then yes", input: "maybe\ny\n", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			p := NewPrompt(in, out)

			got, err := p.Confirm("Continue?")
			if (err != nil) != tt.wantErr {
				t.Errorf("Confirm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Confirm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrompt_Input(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		want         string
		wantErr      bool
	}{
		{name: "simple input", input: "hello\n", want: "hello"},
		{name: "input with spaces", input: "hello world\n", want: "hello world"},
		{name: "empty uses default", input: "\n", defaultValue: "default", want: "default"},
		{name: "empty no default", input: "\n", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			p := NewPrompt(in, out)

			got, err := p.Input("Enter value:", tt.defaultValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Input() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Input() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrompt_SelectWithDefault(t *testing.T) {
	options := []Option{
		{Label: "Option A", Value: "a"},
		{Label: "Option B", Value: "b", Default: true},
		{Label: "Option C", Value: "c"},
	}

	in := strings.NewReader("\n")
	out := &bytes.Buffer{}
	p := NewPrompt(in, out)

	idx, err := p.Select("Choose:", options)
	if err != nil {
		t.Errorf("Select() error = %v", err)
		return
	}
	if idx != 1 {
		t.Errorf("Select() = %v, want 1 (default option)", idx)
	}
}

func TestGuide_Init(t *testing.T) {
	in := strings.NewReader("my-task\n")
	out := &bytes.Buffer{}
	g := New(in, out)

	result, err := g.InitWizard()
	if err != nil {
		t.Errorf("InitWizard() error = %v", err)
		return
	}

	if result.TaskName != "my-task" {
		t.Errorf("TaskName = %v, want my-task", result.TaskName)
	}

	output := out.String()
	if !strings.Contains(output, "Welcome") {
		t.Errorf("output should contain welcome message")
	}
}

func TestGuide_TaskStart(t *testing.T) {
	in := strings.NewReader("fix-bug\nmain\n")
	out := &bytes.Buffer{}
	g := New(in, out)

	result, err := g.TaskStartWizard("main")
	if err != nil {
		t.Errorf("TaskStartWizard() error = %v", err)
		return
	}

	if result.TaskName != "fix-bug" {
		t.Errorf("TaskName = %v, want fix-bug", result.TaskName)
	}
}

func TestGuide_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name      string
		errorCode string
		input     string
		wantCmd   string
	}{
		{
			name:      "not initialized - init",
			errorCode: "not_initialized",
			input:     "1\n",
			wantCmd:   "bar init",
		},
		{
			name:      "not initialized - task start",
			errorCode: "not_initialized",
			input:     "2\nmy-task\n",
			wantCmd:   "bar task start my-task",
		},
		{
			name:      "no active task",
			errorCode: "no_active_task",
			input:     "1\nnew-task\n",
			wantCmd:   "bar task start new-task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}
			g := New(in, out)

			result, err := g.ErrorRecoveryWizard(tt.errorCode)
			if err != nil {
				t.Errorf("ErrorRecoveryWizard() error = %v", err)
				return
			}

			if result.SuggestedCommand != tt.wantCmd {
				t.Errorf("SuggestedCommand = %v, want %v", result.SuggestedCommand, tt.wantCmd)
			}
		})
	}
}

func TestOption_String(t *testing.T) {
	opt := Option{Label: "Test", Value: "test", Description: "A test option"}
	str := opt.String()

	if !strings.Contains(str, "Test") {
		t.Errorf("String() should contain label")
	}
	if !strings.Contains(str, "A test option") {
		t.Errorf("String() should contain description")
	}
}
