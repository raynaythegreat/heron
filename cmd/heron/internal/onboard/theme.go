package onboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/raynaythegreat/heron/cmd/heron/internal"
	"github.com/raynaythegreat/heron/pkg/config"
)

var (
	colorPrimary  = lipgloss.Color("#3E5DB9")
	colorCyan     = lipgloss.Color("#00C8DC")
	colorGreen    = lipgloss.Color("#50C878")
	colorRed      = lipgloss.Color("#D54646")
	colorYellow   = lipgloss.Color("#DCB432")
	colorDim      = lipgloss.Color("#666666")
	colorFg       = lipgloss.Color("#cdd6f4")
	colorMuted    = lipgloss.Color("#6c7086")
	colorSurface0 = lipgloss.Color("#313244")
	colorSurface1 = lipgloss.Color("#45475a")
	colorSurface2 = lipgloss.Color("#585b70")
	colorOverlay0 = lipgloss.Color("#6c7086")
	colorMantle   = lipgloss.Color("#181825")
	colorCrust    = lipgloss.Color("#11111b")
	colorLavender = lipgloss.Color("#b4befe")
	colorBlue     = lipgloss.Color("#89b4fa")
	colorSapphire = lipgloss.Color("#74c7ec")
	colorSky      = lipgloss.Color("#89dceb")
	colorTeal     = lipgloss.Color("#94e2d5")
	colorMauve    = lipgloss.Color("#cba6f7")
	colorPink     = lipgloss.Color("#f5c2e7")
	colorPeach    = lipgloss.Color("#fab387")
	colorMaroon   = lipgloss.Color("#eba0ac")
	colorRose     = lipgloss.Color("#f38ba8")
	colorFlamingo = lipgloss.Color("#f2cdcd")
)

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorLavender).
			MarginBottom(0)

	styleSubtitle = lipgloss.NewStyle().
			Foreground(colorOverlay0).
			MarginBottom(0)

	styleDivider = lipgloss.NewStyle().
			Foreground(colorSurface1).
			Render("────────────────────────────────────────────")

	styleStepHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue).
			MarginBottom(0)

	styleProgress = lipgloss.NewStyle().
			Foreground(colorMauve).
			Bold(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorGreen)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed)

	styleWarning = lipgloss.NewStyle().
			Foreground(colorYellow)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDim)

	styleHelp = lipgloss.NewStyle().
			Foreground(colorOverlay0).
			Italic(true)

	styleBody = lipgloss.NewStyle().
			Foreground(colorFg).
			MarginBottom(0)

	styleDone = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGreen)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSurface2).
			Padding(1, 2)

	boxAccentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMauve).
			Padding(1, 2)

	boxSuccessStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorGreen).
			Padding(1, 3)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorOverlay0).
			Width(16)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorFg).
			Bold(true)
)

func stepHeader(step, total int, title string) string {
	pbar := renderProgressBar(step, total)
	header := styleStepHeader.Render(title)
	return pbar + "\n" + header
}

func renderProgressBar(current, total int) string {
	width := 40
	filled := (current * width) / total
	empty := width - filled

	coloredBar := lipgloss.NewStyle().Foreground(colorMauve).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(colorSurface1).Render(strings.Repeat("░", empty))

	label := fmt.Sprintf(" %d/%d ", current, total)
	labelS := lipgloss.NewStyle().Foreground(colorOverlay0).Bold(true)

	return coloredBar + labelS.Render(label)
}

func formatSuccess(msg string) string {
	return styleSuccess.Render("  ✓ " + msg)
}

func formatError(msg string) string {
	return styleError.Render("  ✗ " + msg)
}

func formatWarning(msg string) string {
	return styleWarning.Render("  ⚠ " + msg)
}

func formatDim(msg string) string {
	return styleDim.Render(msg)
}

func formatHelp(msg string) string {
	return styleHelp.Render(msg)
}

func formatBody(msg string) string {
	return styleBody.Render(msg)
}

func formatDone(msg string) string {
	return styleDone.Render(msg)
}

func renderWelcomeScreen() string {
	logo := lipgloss.NewStyle().Foreground(colorMauve).Bold(true).Render(internal.Logo)
	title := lipgloss.NewStyle().
		Foreground(colorLavender).
		Bold(true).
		Render("Setup Wizard")

	subtitle := lipgloss.NewStyle().
		Foreground(colorOverlay0).
		Render("Configure your AI providers, models, and tools")

	reRun := lipgloss.NewStyle().
		Foreground(colorSurface2).
		Render("Re-run anytime with: ") +
		lipgloss.NewStyle().Foreground(colorSky).Render("heron onboard")

	content := fmt.Sprintf("%s  %s\n\n%s\n\n%s", logo, title, subtitle, reRun)
	return boxAccentStyle.Render(content)
}

func renderCompletionSummary(cfg *config.Config, configPath string) string {
	var lines []string

	lines = append(lines,
		lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("  ✓  Heron is configured and ready!"),
		"",
	)

	if cfg.Agents.Defaults.ModelName != "" {
		lines = append(lines,
			labelStyle.Render("Primary Model")+"  "+valueStyle.Render(cfg.Agents.Defaults.ModelName),
		)
	}
	if len(cfg.Agents.Defaults.ModelFallbacks) > 0 {
		lines = append(lines,
			labelStyle.Render("Fallbacks")+"  "+lipgloss.NewStyle().Foreground(colorOverlay0).Render(strings.Join(cfg.Agents.Defaults.ModelFallbacks, " → ")),
		)
	}
	if cfg.Agents.Defaults.Routing != nil && cfg.Agents.Defaults.Routing.Enabled {
		lines = append(lines,
			labelStyle.Render("Smart Routing")+"  "+styleSuccess.Render("enabled")+
				lipgloss.NewStyle().Foreground(colorOverlay0).Render(fmt.Sprintf(" (light: %s)", cfg.Agents.Defaults.Routing.LightModel)),
		)
	}

	lines = append(lines,
		"",
		labelStyle.Render("Config")+"  "+lipgloss.NewStyle().Foreground(colorDim).Render(configPath),
		"",
		lipgloss.NewStyle().Foreground(colorOverlay0).Bold(true).Render("  Next steps"),
		"",
		lipgloss.NewStyle().Foreground(colorSky).Render("  ● heron agent -m \"Hello!\"")+"     "+lipgloss.NewStyle().Foreground(colorDim).Render("Chat via CLI"),
		lipgloss.NewStyle().Foreground(colorSky).Render("  ● heron web --console")+"      "+lipgloss.NewStyle().Foreground(colorDim).Render("Web UI → localhost:18800"),
		lipgloss.NewStyle().Foreground(colorSky).Render("  ● heron-launcher")+"          "+lipgloss.NewStyle().Foreground(colorDim).Render("Terminal UI"),
	)

	return boxSuccessStyle.Render(strings.Join(lines, "\n"))
}

func heronTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Form.Base = t.Form.Base.Padding(0, 1)

	t.Group.Base = t.Group.Base.Margin(1, 0)
	t.Group.Title = lipgloss.NewStyle().
		Foreground(colorBlue).
		Bold(true).
		MarginBottom(0)
	t.Group.Description = lipgloss.NewStyle().
		Foreground(colorOverlay0).
		MarginBottom(0)

	t.Focused.Base = lipgloss.NewStyle().Padding(0, 1)
	t.Focused.Title = lipgloss.NewStyle().
		Foreground(colorLavender).
		Bold(true)
	t.Focused.Description = lipgloss.NewStyle().
		Foreground(colorOverlay0)

	t.Focused.SelectSelector = lipgloss.NewStyle().
		Foreground(colorMauve).
		Bold(true)
	t.Focused.Option = lipgloss.NewStyle().
		Foreground(colorFg)
	t.Focused.NextIndicator = lipgloss.NewStyle().
		Foreground(colorMauve)
	t.Focused.PrevIndicator = lipgloss.NewStyle().
		Foreground(colorMauve)

	t.Focused.MultiSelectSelector = lipgloss.NewStyle().
		Foreground(colorMauve).
		Bold(true)
	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(colorGreen)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		Foreground(colorGreen).
		Bold(true)
	t.Focused.UnselectedOption = lipgloss.NewStyle().
		Foreground(colorFg)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		Foreground(colorSurface2)

	t.Focused.FocusedButton = lipgloss.NewStyle().
		Foreground(colorCrust).
		Background(colorMauve).
		Bold(true).
		Padding(0, 2)
	t.Focused.BlurredButton = lipgloss.NewStyle().
		Foreground(colorOverlay0).
		Padding(0, 2)

	t.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(colorMauve)
	t.Focused.TextInput.Text = lipgloss.NewStyle().
		Foreground(colorFg)
	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(colorSurface2)
	t.Focused.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(colorMauve).
		Bold(true)

	t.Focused.ErrorIndicator = lipgloss.NewStyle().
		Foreground(colorRed)
	t.Focused.ErrorMessage = lipgloss.NewStyle().
		Foreground(colorRed)

	t.Focused.Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSurface2).
		Padding(0, 1)
	t.Focused.NoteTitle = lipgloss.NewStyle().
		Foreground(colorLavender).
		Bold(true)
	t.Focused.Next = lipgloss.NewStyle().
		Foreground(colorMauve).
		Bold(true)

	t.Blurred.Title = lipgloss.NewStyle().
		Foreground(colorOverlay0)
	t.Blurred.Description = lipgloss.NewStyle().
		Foreground(colorDim)
	t.Blurred.SelectedOption = lipgloss.NewStyle().
		Foreground(colorGreen)
	t.Blurred.SelectedPrefix = lipgloss.NewStyle().
		Foreground(colorGreen)
	t.Blurred.UnselectedOption = lipgloss.NewStyle().
		Foreground(colorDim)
	t.Blurred.UnselectedPrefix = lipgloss.NewStyle().
		Foreground(colorSurface2)

	t.FieldSeparator = lipgloss.NewStyle().
		Foreground(colorSurface1)

	return t
}
