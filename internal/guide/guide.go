package guide

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Option struct {
	Label       string
	Value       string
	Description string
	Default     bool
}

func (o Option) String() string {
	if o.Description != "" {
		return fmt.Sprintf("%s - %s", o.Label, o.Description)
	}
	return o.Label
}

type Prompt struct {
	reader *bufio.Reader
	writer io.Writer
}

func NewPrompt(in io.Reader, out io.Writer) *Prompt {
	return &Prompt{
		reader: bufio.NewReader(in),
		writer: out,
	}
}

func (p *Prompt) Select(message string, options []Option) (int, error) {
	defaultIdx := -1
	for i, opt := range options {
		if opt.Default {
			defaultIdx = i
			break
		}
	}

	for {
		fmt.Fprintln(p.writer, message)
		for i, opt := range options {
			marker := "  "
			if opt.Default {
				marker = "* "
			}
			fmt.Fprintf(p.writer, "%s%d) %s\n", marker, i+1, opt.String())
		}

		if defaultIdx >= 0 {
			fmt.Fprintf(p.writer, "Enter choice [%d]: ", defaultIdx+1)
		} else {
			fmt.Fprint(p.writer, "Enter choice: ")
		}

		line, err := p.reader.ReadString('\n')
		if err != nil {
			return 0, err
		}

		line = strings.TrimSpace(line)
		if line == "" && defaultIdx >= 0 {
			return defaultIdx, nil
		}

		num, err := strconv.Atoi(line)
		if err != nil {
			fmt.Fprintln(p.writer, "Invalid input. Please enter a number.")
			continue
		}

		if num < 1 || num > len(options) {
			fmt.Fprintf(p.writer, "Please enter a number between 1 and %d.\n", len(options))
			continue
		}

		return num - 1, nil
	}
}

func (p *Prompt) Confirm(message string) (bool, error) {
	for {
		fmt.Fprintf(p.writer, "%s [y/N]: ", message)

		line, err := p.reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		line = strings.TrimSpace(strings.ToLower(line))
		switch line {
		case "y", "yes":
			return true, nil
		case "n", "no", "":
			return false, nil
		default:
			fmt.Fprintln(p.writer, "Please enter 'y' or 'n'.")
		}
	}
}

func (p *Prompt) Input(message string, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.writer, "%s [%s]: ", message, defaultValue)
	} else {
		fmt.Fprintf(p.writer, "%s ", message)
	}

	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue, nil
	}
	return line, nil
}

type Guide struct {
	prompt *Prompt
	writer io.Writer
}

func New(in io.Reader, out io.Writer) *Guide {
	return &Guide{
		prompt: NewPrompt(in, out),
		writer: out,
	}
}

type InitResult struct {
	TaskName string
	BaseRef  string
}

func (g *Guide) InitWizard() (*InitResult, error) {
	fmt.Fprintln(g.writer, "")
	fmt.Fprintln(g.writer, "🚀 Welcome to BAR (Blade Agent Runtime)!")
	fmt.Fprintln(g.writer, "")
	fmt.Fprintln(g.writer, "BAR helps you manage AI agent tasks in isolated workspaces.")
	fmt.Fprintln(g.writer, "")

	taskName, err := g.prompt.Input("Enter a name for your first task:", "")
	if err != nil {
		return nil, err
	}

	return &InitResult{
		TaskName: taskName,
	}, nil
}

type TaskStartResult struct {
	TaskName string
	BaseRef  string
}

func (g *Guide) TaskStartWizard(defaultBase string) (*TaskStartResult, error) {
	fmt.Fprintln(g.writer, "")
	fmt.Fprintln(g.writer, "📋 Create a new task")
	fmt.Fprintln(g.writer, "")

	taskName, err := g.prompt.Input("Task name:", "")
	if err != nil {
		return nil, err
	}

	baseRef, err := g.prompt.Input("Base branch/commit:", defaultBase)
	if err != nil {
		return nil, err
	}

	return &TaskStartResult{
		TaskName: taskName,
		BaseRef:  baseRef,
	}, nil
}

type ErrorRecoveryResult struct {
	SuggestedCommand string
	AutoExecute      bool
}

func (g *Guide) ErrorRecoveryWizard(errorCode string) (*ErrorRecoveryResult, error) {
	switch errorCode {
	case "not_initialized":
		return g.notInitializedRecovery()
	case "no_active_task":
		return g.noActiveTaskRecovery()
	default:
		return &ErrorRecoveryResult{}, nil
	}
}

func (g *Guide) notInitializedRecovery() (*ErrorRecoveryResult, error) {
	fmt.Fprintln(g.writer, "")
	fmt.Fprintln(g.writer, "BAR is not initialized in this repository.")
	fmt.Fprintln(g.writer, "")

	options := []Option{
		{Label: "Initialize BAR only", Value: "init", Description: "Run 'bar init'"},
		{Label: "Initialize and create a task", Value: "task", Description: "Run 'bar task start'", Default: true},
		{Label: "Cancel", Value: "cancel"},
	}

	idx, err := g.prompt.Select("What would you like to do?", options)
	if err != nil {
		return nil, err
	}

	switch options[idx].Value {
	case "init":
		return &ErrorRecoveryResult{
			SuggestedCommand: "bar init",
			AutoExecute:      true,
		}, nil
	case "task":
		taskName, err := g.prompt.Input("Task name:", "")
		if err != nil {
			return nil, err
		}
		return &ErrorRecoveryResult{
			SuggestedCommand: fmt.Sprintf("bar task start %s", taskName),
			AutoExecute:      true,
		}, nil
	default:
		return &ErrorRecoveryResult{}, nil
	}
}

func (g *Guide) noActiveTaskRecovery() (*ErrorRecoveryResult, error) {
	fmt.Fprintln(g.writer, "")
	fmt.Fprintln(g.writer, "No active task found.")
	fmt.Fprintln(g.writer, "")

	options := []Option{
		{Label: "Create a new task", Value: "new", Default: true},
		{Label: "Switch to an existing task", Value: "switch"},
		{Label: "Cancel", Value: "cancel"},
	}

	idx, err := g.prompt.Select("What would you like to do?", options)
	if err != nil {
		return nil, err
	}

	switch options[idx].Value {
	case "new":
		taskName, err := g.prompt.Input("Task name:", "")
		if err != nil {
			return nil, err
		}
		return &ErrorRecoveryResult{
			SuggestedCommand: fmt.Sprintf("bar task start %s", taskName),
			AutoExecute:      true,
		}, nil
	case "switch":
		return &ErrorRecoveryResult{
			SuggestedCommand: "bar task list",
			AutoExecute:      true,
		}, nil
	default:
		return &ErrorRecoveryResult{}, nil
	}
}

func (g *Guide) Print(message string) {
	fmt.Fprintln(g.writer, message)
}

func (g *Guide) Prompt() *Prompt {
	return g.prompt
}

func (g *Guide) Printf(format string, args ...any) {
	fmt.Fprintf(g.writer, format, args...)
}
