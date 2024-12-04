package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderStats() string {
	runtime := time.Since(m.startTime).Round(time.Second)

	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("21")).
		Width(m.viewport.Width).
		Padding(0, 0)

	valueStyle := lipgloss.NewStyle()

	separator := lipgloss.NewStyle().
		SetString(" │ ").
		String()

	stats := baseStyle.Render(fmt.Sprintf("SPS: %s%sSkeets: %s%sFound: %s%sTotal: %s%sRuntime: %s",
		valueStyle.Render(fmt.Sprintf("%6.2f", m.stats.Rate)), separator,
		valueStyle.Render(fmt.Sprintf("%3d", m.stats.Posts)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.Amulets)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.TotalAmulets)), separator,
		valueStyle.Render(runtime.String())))

	return stats
}

func (m Model) renderEntries() string {
	var output strings.Builder

	normalRow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Width(m.viewport.Width).
		Padding(0, 0)

	alternateRow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("234")).
		Width(m.viewport.Width).
		Padding(0, 0)

	for i, e := range m.entries {
		style := normalRow
		if i%2 == 0 {
			style = alternateRow
		}

		coloredText := style.Render(e.Text)
		output.WriteString(coloredText + "\n")
	}

	return output.String()
}
