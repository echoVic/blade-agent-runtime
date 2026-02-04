package ui

import (
	"strings"
	"testing"
)

func TestBox_Render(t *testing.T) {
	box := NewBox("🔧 BAR Status")
	box.AddRow("Repository", "/path/to/repo")
	box.AddRow("Active Task", "fix-bug (abc123)")
	box.AddRow("Branch", "bar/fix-bug-abc123")
	box.AddRow("Status", "🟡 dirty (3 files)")
	box.AddRow("Steps", "5")

	output := box.Render()

	if !strings.Contains(output, "╭") {
		t.Error("output should contain top-left corner")
	}
	if !strings.Contains(output, "╰") {
		t.Error("output should contain bottom-left corner")
	}
	if !strings.Contains(output, "🔧 BAR Status") {
		t.Error("output should contain title")
	}
	if !strings.Contains(output, "Repository") {
		t.Error("output should contain Repository label")
	}
	if !strings.Contains(output, "/path/to/repo") {
		t.Error("output should contain Repository value")
	}
	if !strings.Contains(output, "fix-bug (abc123)") {
		t.Error("output should contain task info")
	}
}

func TestBox_EmptyTitle(t *testing.T) {
	box := NewBox("")
	box.AddRow("Key", "Value")

	output := box.Render()

	if strings.Contains(output, "├") {
		t.Error("output should not contain separator when no title")
	}
	if !strings.Contains(output, "Key") {
		t.Error("output should contain row")
	}
}

func TestBox_Width(t *testing.T) {
	box := NewBox("Title")
	box.SetWidth(30)
	box.AddRow("A", "B")

	output := box.Render()
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		runeCount := runeWidth(line)
		if runeCount != 30 {
			t.Errorf("line width = %d, want 30: %q", runeCount, line)
		}
	}
}

func TestBox_AutoWidth(t *testing.T) {
	box := NewBox("Status")
	box.AddRow("Short", "A")
	box.AddRow("Very Long Label Here", "Some very long value that should expand the box")

	output := box.Render()
	lines := strings.Split(output, "\n")

	width := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		w := runeWidth(line)
		if width == 0 {
			width = w
		}
		if w != width {
			t.Errorf("inconsistent width: got %d, want %d", w, width)
		}
	}
}

func TestBox_StatusIndicator(t *testing.T) {
	tests := []struct {
		name   string
		clean  bool
		files  int
		want   string
	}{
		{"clean", true, 0, "🟢 clean"},
		{"dirty with files", false, 3, "🟡 dirty (3 files)"},
		{"dirty no files", false, 0, "🟡 dirty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusIndicator(tt.clean, tt.files)
			if got != tt.want {
				t.Errorf("StatusIndicator() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBox_Padding(t *testing.T) {
	box := NewBox("Test")
	box.SetPadding(2)
	box.AddRow("Key", "Value")

	output := box.Render()

	if !strings.Contains(output, "  Key") {
		t.Error("output should have left padding")
	}
}

func runeWidth(s string) int {
	count := 0
	for _, r := range s {
		if r >= 0x1100 && (r <= 0x115F || r == 0x2329 || r == 0x232A ||
			(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFE10 && r <= 0xFE1F) ||
			(r >= 0xFE30 && r <= 0xFE6F) ||
			(r >= 0xFF00 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6) ||
			(r >= 0x1F300 && r <= 0x1F9FF)) {
			count += 2
		} else {
			count++
		}
	}
	return count
}
