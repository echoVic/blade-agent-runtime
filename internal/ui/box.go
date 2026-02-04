package ui

import (
	"fmt"
	"strings"
)

type Row struct {
	Label string
	Value string
}

type Box struct {
	title   string
	rows    []Row
	width   int
	padding int
}

func NewBox(title string) *Box {
	return &Box{
		title:   title,
		rows:    make([]Row, 0),
		width:   0,
		padding: 1,
	}
}

func (b *Box) AddRow(label, value string) {
	b.rows = append(b.rows, Row{Label: label, Value: value})
}

func (b *Box) SetWidth(w int) {
	b.width = w
}

func (b *Box) SetPadding(p int) {
	b.padding = p
}

func (b *Box) Render() string {
	width := b.calculateWidth()
	innerWidth := width - 2

	var sb strings.Builder

	sb.WriteString("╭")
	sb.WriteString(strings.Repeat("─", innerWidth))
	sb.WriteString("╮\n")

	if b.title != "" {
		titleLine := b.padRight(fmt.Sprintf("%s%s", strings.Repeat(" ", b.padding), b.title), innerWidth)
		sb.WriteString("│")
		sb.WriteString(titleLine)
		sb.WriteString("│\n")

		sb.WriteString("├")
		sb.WriteString(strings.Repeat("─", innerWidth))
		sb.WriteString("┤\n")
	}

	labelWidth := b.maxLabelWidth()
	for _, row := range b.rows {
		content := fmt.Sprintf("%s%-*s  %s",
			strings.Repeat(" ", b.padding),
			labelWidth,
			row.Label,
			row.Value,
		)
		line := b.padRight(content, innerWidth)
		sb.WriteString("│")
		sb.WriteString(line)
		sb.WriteString("│\n")
	}

	sb.WriteString("╰")
	sb.WriteString(strings.Repeat("─", innerWidth))
	sb.WriteString("╯")

	return sb.String()
}

func (b *Box) calculateWidth() int {
	if b.width > 0 {
		return b.width
	}

	maxWidth := runeLen(b.title) + b.padding*2 + 2

	labelWidth := b.maxLabelWidth()
	for _, row := range b.rows {
		rowWidth := b.padding + labelWidth + 2 + runeLen(row.Value) + b.padding + 2
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}

	return maxWidth
}

func (b *Box) maxLabelWidth() int {
	max := 0
	for _, row := range b.rows {
		l := runeLen(row.Label)
		if l > max {
			max = l
		}
	}
	return max
}

func (b *Box) padRight(s string, width int) string {
	currentWidth := runeLen(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

func runeLen(s string) int {
	count := 0
	for _, r := range s {
		if isWideRune(r) {
			count += 2
		} else {
			count++
		}
	}
	return count
}

func isWideRune(r rune) bool {
	return r >= 0x1100 && (r <= 0x115F || r == 0x2329 || r == 0x232A ||
		(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE10 && r <= 0xFE1F) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x1F300 && r <= 0x1F9FF))
}

func StatusIndicator(clean bool, files int) string {
	if clean {
		return "🟢 clean"
	}
	if files > 0 {
		return fmt.Sprintf("🟡 dirty (%d files)", files)
	}
	return "🟡 dirty"
}
