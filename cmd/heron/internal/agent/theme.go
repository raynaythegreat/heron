package agent

import (
	"os"
	"strconv"
	"strings"
)

type Theme struct {
	Colors Colors
}

type Colors struct {
	Background  int32
	Surface     int32
	SelectedBg  int32
	Border      int32
	Dim         int32
	Accent      int32
	Text        int32
	Error       int32
	Success     int32
	Quote       int32
	QuoteBorder int32
	Code        int32
	CodeBlock   int32
	Link        int32
	ToolPending int32
	ToolSuccess int32
	ToolError   int32
	ToolTitle   int32
	ToolOutput  int32
	UserBg      int32
	UserText    int32
	SystemText  int32
	ModePlan    int32
	ModeBuild   int32
	ModeChat    int32
	Spinner     int32
	Label       int32
	InputBg     int32
	InputText   int32
	InputBorder int32
	FooterBg    int32
}

func hexToInt(hex string) int32 {
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return int32((r << 16) | (g << 8) | b)
}

func darkPalette() Colors {
	return Colors{
		Background:  hexToInt("#0A0A12"),
		Surface:     hexToInt("#12101F"),
		SelectedBg:  hexToInt("#1E0F3D"),
		Border:      hexToInt("#3C414B"),
		Dim:         hexToInt("#7B7F87"),
		Accent:      hexToInt("#06B6D4"),
		Text:        hexToInt("#E8E3D5"),
		Error:       hexToInt("#F97066"),
		Success:     hexToInt("#7DD3A5"),
		Quote:       hexToInt("#8CC8FF"),
		QuoteBorder: hexToInt("#3B4D6B"),
		Code:        hexToInt("#F0C987"),
		CodeBlock:   hexToInt("#1E232A"),
		Link:        hexToInt("#7DD3A5"),
		ToolPending: hexToInt("#1F2A2F"),
		ToolSuccess: hexToInt("#1E2D23"),
		ToolError:   hexToInt("#2F1F1F"),
		ToolTitle:   hexToInt("#06B6D4"),
		ToolOutput:  hexToInt("#E8E3D5"),
		UserBg:      hexToInt("#2B2F36"),
		UserText:    hexToInt("#F3EEE0"),
		SystemText:  hexToInt("#9BA3B2"),
		ModePlan:    hexToInt("#F59E0B"),
		ModeBuild:   hexToInt("#7DD3A5"),
		ModeChat:    hexToInt("#06B6D4"),
		Spinner:     hexToInt("#06B6D4"),
		Label:       hexToInt("#06B6D4"),
		InputBg:     hexToInt("#0A0A12"),
		InputText:   hexToInt("#E8E3D5"),
		InputBorder: hexToInt("#3C414B"),
		FooterBg:    hexToInt("#12101F"),
	}
}

func lightPalette() Colors {
	return Colors{
		Background:  hexToInt("#FFFFFF"),
		Surface:     hexToInt("#E5E2DA"),
		SelectedBg:  hexToInt("#E0F2FE"),
		Border:      hexToInt("#5B6472"),
		Dim:         hexToInt("#5B6472"),
		Accent:      hexToInt("#0891B2"),
		Text:        hexToInt("#1E1E1E"),
		Error:       hexToInt("#DC2626"),
		Success:     hexToInt("#047857"),
		Quote:       hexToInt("#1D4ED8"),
		QuoteBorder: hexToInt("#2563EB"),
		Code:        hexToInt("#92400E"),
		CodeBlock:   hexToInt("#F9FAFB"),
		Link:        hexToInt("#047857"),
		ToolPending: hexToInt("#EFF6FF"),
		ToolSuccess: hexToInt("#ECFDF5"),
		ToolError:   hexToInt("#FEF2F2"),
		ToolTitle:   hexToInt("#0891B2"),
		ToolOutput:  hexToInt("#374151"),
		UserBg:      hexToInt("#F3F0E8"),
		UserText:    hexToInt("#1E1E1E"),
		SystemText:  hexToInt("#4B5563"),
		ModePlan:    hexToInt("#D97706"),
		ModeBuild:   hexToInt("#047857"),
		ModeChat:    hexToInt("#0891B2"),
		Spinner:     hexToInt("#0891B2"),
		Label:       hexToInt("#0891B2"),
		InputBg:     hexToInt("#FFFFFF"),
		InputText:   hexToInt("#1E1E1E"),
		InputBorder: hexToInt("#5B6472"),
		FooterBg:    hexToInt("#E5E2DA"),
	}
}

func parseColorFGBG(val string) bool {
	parts := strings.Split(val, ";")
	if len(parts) < 2 {
		return false
	}
	bg := parts[len(parts)-1]
	switch bg {
	case "7", "15":
		return true
	default:
		n, err := strconv.Atoi(bg)
		return err == nil && n >= 244
	}
}

func DetectTheme() *Theme {
	explicit := os.Getenv("HERON_THEME")
	switch strings.ToLower(explicit) {
	case "light":
		return &Theme{Colors: lightPalette()}
	case "dark":
		return &Theme{Colors: darkPalette()}
	}

	if colorfgbg := os.Getenv("COLORFGBG"); colorfgbg != "" {
		if parseColorFGBG(colorfgbg) {
			return &Theme{Colors: lightPalette()}
		}
		return &Theme{Colors: darkPalette()}
	}

	return &Theme{Colors: darkPalette()}
}

var theme *Theme

func GetTheme() *Theme {
	if theme == nil {
		theme = DetectTheme()
	}
	return theme
}
