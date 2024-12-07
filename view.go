package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	containerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("0")).
		Height(m.viewport.Height + 1)

	return containerStyle.Render(fmt.Sprintf("%s\n%s",
		m.renderStats(),
		m.viewport.View()))
}

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

	if color, exists := rarityColors[minRarity]; exists {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Render("•")
	}
	return "?"
}

func (m Model) renderStats() string {
	runtime := time.Since(m.startTime).Round(time.Second)

	baseStyle := lipgloss.NewStyle()
	valueStyle := lipgloss.NewStyle()

	separator := lipgloss.NewStyle().
		SetString(" │ ").
		String()

	stats := baseStyle.Render(fmt.Sprintf("SPS: %s%sSkeets: %s%sNew: %s%sTotal: %s%s%s",
		valueStyle.Render(fmt.Sprintf("%6.2f", m.stats.Rate)), separator,
		valueStyle.Render(fmt.Sprintf("%3d", m.stats.Posts)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.Amulets)), separator,
		valueStyle.Render(fmt.Sprintf("%d", m.stats.TotalAmulets)), separator,
		valueStyle.Render(runtime.String())))

	return stats
}

func (m Model) renderEntries() string {
	var output strings.Builder

	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(m.viewport.Width - 11)

	for _, e := range m.entries {
		symbol := getRaritySymbol(e.Rarity)
		timeStr := e.Time.Format("15:04:05")
		formattedTime := timeStyle.Render(timeStr)
		wrappedText := textStyle.Render(e.Text)

		lines := strings.Split(wrappedText, "\n")

		firstLine := fmt.Sprintf("%s %s %s", formattedTime, symbol, lines[0])
		output.WriteString(firstLine + "\n")

		if len(lines) > 1 {
			padding := strings.Repeat(" ", 11)
			for _, line := range lines[1:] {
				output.WriteString(padding + line + "\n")
			}
		}
	}

	return output.String()
}
