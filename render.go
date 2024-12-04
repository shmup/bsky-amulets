package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func getRaritySymbol(minRarity int) string {
	rarityColors := map[int]string{
		1: "173", // copper/bronze
		2: "247", // silver
		3: "220", // gold
		4: "199", // epic (purple)
		5: "39",  // legendary (green)
		6: "201", // mythic (pink)
		7: "252", // unknown
	}

	raritySymbols := map[int]string{
		1: "C",
		2: "U",
		3: "R",
		4: "E",
		5: "L",
		6: "M",
		7: "?",
	}

	if color, exists := rarityColors[minRarity]; exists {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Render(raritySymbols[minRarity])
	}
	return "?"
}

func (m Model) renderStats() string {
	runtime := time.Since(m.startTime).Round(time.Second)

	baseStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("18")).
		Width(m.viewport.Width).
		Padding(0, 0)

	valueStyle := lipgloss.NewStyle()

	separator := lipgloss.NewStyle().
		SetString(" │ ").
		String()

	stats := baseStyle.Render(fmt.Sprintf("SPS: %s%sSkeets: %s%sNew: %s%sTotal: %s%s%s%s",
		valueStyle.Render(fmt.Sprintf("%6.2f", m.stats.Rate)), separator,
		valueStyle.Render(fmt.Sprintf("%3d", m.stats.Posts)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.Amulets)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.TotalAmulets)), separator,
		valueStyle.Render(runtime.String()), separator))

	return stats
}

func (m Model) renderEntries() string {
	var output strings.Builder

	rowStyle := lipgloss.NewStyle().
		Width(m.viewport.Width).
		Padding(0, 0)

	for _, e := range m.entries {
		symbol := getRaritySymbol(e.Rarity - 3)
		styledText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Render(e.Text)

		line := fmt.Sprintf("%s %s", symbol, styledText)
		coloredLine := rowStyle.Render(line)
		output.WriteString(coloredLine + "\n")
	}

	return output.String()
}
