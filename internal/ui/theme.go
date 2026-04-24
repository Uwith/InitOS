package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	StartColor = "#FF69B4"
	EndColor   = "#800080"
)

type Theme struct {
	Header       lipgloss.Style
	Subtle       lipgloss.Style
	Normal       lipgloss.Style
	Cursor       lipgloss.Style
	Selected     lipgloss.Style
	Installed    lipgloss.Style
	FilterPrompt lipgloss.Style
	Help         lipgloss.Style
	Success      lipgloss.Style
	Failure      lipgloss.Style
}

func NewTheme() Theme {
	return Theme{
		Header:       lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Subtle:       lipgloss.NewStyle().Foreground(lipgloss.Color("#A8A8A8")),
		Normal:       lipgloss.NewStyle().Foreground(lipgloss.Color("#EDEDED")),
		Cursor:       lipgloss.NewStyle().Foreground(lipgloss.Color(StartColor)).Bold(true),
		Selected:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8CC6")).Bold(true),
		Installed:    lipgloss.NewStyle().Foreground(lipgloss.Color("#60D394")),
		FilterPrompt: lipgloss.NewStyle().Foreground(lipgloss.Color("#D084E8")).Bold(true),
		Help:         lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")),
		Success:      lipgloss.NewStyle().Foreground(lipgloss.Color("#62D26F")).Bold(true),
		Failure:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5C7A")).Bold(true),
	}
}

func GradientText(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) == 1 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(StartColor)).Render(text)
	}

	var b strings.Builder
	for i, r := range runes {
		t := float64(i) / float64(len(runes)-1)
		color := mixColor(StartColor, EndColor, t)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(string(r)))
	}
	return b.String()
}

func mixColor(start, end string, ratio float64) string {
	sr, sg, sb := parseHex(start)
	er, eg, eb := parseHex(end)
	r := int(float64(sr) + (float64(er)-float64(sr))*ratio)
	g := int(float64(sg) + (float64(eg)-float64(sg))*ratio)
	b := int(float64(sb) + (float64(eb)-float64(sb))*ratio)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func parseHex(hex string) (int, int, int) {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
