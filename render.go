package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func getRarityCircle(minRarity int) string {
	rarityColors := map[int]string{
		1: "173", // copper/bronze
		2: "247", // silver
		3: "220", // gold
		4: "199", // epic (purple)
		5: "39",  // legendary (green)
		6: "201", // mythic (pink)
	}

	// default to question mark for unknown rarities
	if color, exists := rarityColors[minRarity]; exists {
		return lipgloss.NewStyle().
			SetString("●").
			Foreground(lipgloss.Color(color)).
			String()
	}
	return "?"
}

func (m Model) renderStats(minRarity *int) string {
	runtime := time.Since(m.startTime).Round(time.Second)

	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("21")).
		Width(m.viewport.Width).
		Padding(0, 0)

	valueStyle := lipgloss.NewStyle()

	separator := lipgloss.NewStyle().
		SetString(" │ ").
		String()

	rarityIndicator := getRarityCircle(*minRarity)

	stats := baseStyle.Render(fmt.Sprintf("SPS: %s%sSkeets: %s%sNew: %s%sTotal: %s%s%s%s%s",
		valueStyle.Render(fmt.Sprintf("%6.2f", m.stats.Rate)), separator,
		valueStyle.Render(fmt.Sprintf("%3d", m.stats.Posts)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.Amulets)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.TotalAmulets)), separator,
		valueStyle.Render(runtime.String()),
		separator, rarityIndicator))

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
