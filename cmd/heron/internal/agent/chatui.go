package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/raynaythegreat/heron/cmd/heron/internal"
	pkgagent "github.com/raynaythegreat/heron/pkg/agent"
	pkgconfig "github.com/raynaythegreat/heron/pkg/config"
	"github.com/raynaythegreat/heron/pkg/providers"
)

const cardWidth = 50

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var waitingPhrases = []string{
	"thinking",
	"analyzing",
	"reasoning",
	"processing",
	"searching knowledge",
	"consulting instincts",
	"planning execution",
	"crafting response",
	"summarizing findings",
	"formulating answer",
	"exploring options",
	"connecting dots",
	"preparing response",
	"refining output",
	"organizing thoughts",
}

type cmdSuggestion struct {
	cmd  string
	desc string
}

type modelChoice struct {
	name    string
	modelID string
}

// chatMode controls how the agent approaches responses.
type chatMode int

const (
	modeBuild chatMode = iota // default: direct execution
	modePlan                  // plan before acting
	modeChat                  // conversational/research mode
)

func (m chatMode) String() string {
	switch m {
	case modePlan:
		return "PLAN"
	case modeChat:
		return "CHAT"
	default:
		return "BUILD"
	}
}

type chatUI struct {
	app    *tview.Application
	pages  *tview.Pages
	layout *tview.Flex

	header     *tview.TextView
	chatLog    *tview.TextView
	statusLine *tview.TextView
	footer     *tview.TextView
	input      *tview.InputField

	// slash-command autocomplete
	suggList    *tview.List
	suggVisible bool
	allSugg     []cmdSuggestion

	modelName  string
	sessionKey string
	agentLoop  *pkgagent.AgentLoop
	mode       chatMode // Plan vs Build

	mu         sync.Mutex
	busy       bool
	spinIdx    int
	startTime  time.Time
	lastTool   string
	history    []string
	histIdx    int
	ctx        context.Context // set during run(), used by sendToLoop
	skillNames map[string]bool // lowercase skill name → true
	tempStatus string          // temporary status for UI feedback
}

func newChatUI(modelName, sessionKey string, agentLoop *pkgagent.AgentLoop) *chatUI {
	c := &chatUI{
		modelName:  modelName,
		sessionKey: sessionKey,
		agentLoop:  agentLoop,
		mode:       modeChat,
	}
	c.buildLayout()
	return c
}

func (c *chatUI) shortSession() string {
	if len(c.sessionKey) > 8 {
		return c.sessionKey[:8]
	}
	return c.sessionKey
}

func (c *chatUI) buildLayout() {
	t := GetTheme()
	cl := t.Colors

	c.header = tview.NewTextView().SetDynamicColors(true)
	c.header.SetBackgroundColor(tcell.NewHexColor(cl.Background))

	c.chatLog = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true)
	c.chatLog.SetBackgroundColor(tcell.NewHexColor(cl.Background))
	c.chatLog.SetTextColor(tcell.NewHexColor(cl.Text))

	c.statusLine = tview.NewTextView().SetDynamicColors(true)
	c.statusLine.SetBackgroundColor(tcell.NewHexColor(cl.Surface))
	c.statusLine.SetText(fmt.Sprintf("  [#%06X]idle[-]", cl.Dim))

	c.footer = tview.NewTextView().SetDynamicColors(true)
	c.footer.SetBackgroundColor(tcell.NewHexColor(cl.Surface))

	c.suggList = tview.NewList()
	c.suggList.ShowSecondaryText(true)
	c.suggList.SetBackgroundColor(tcell.NewHexColor(cl.Surface))
	c.suggList.SetMainTextColor(tcell.NewHexColor(cl.Text))
	c.suggList.SetSecondaryTextColor(tcell.NewHexColor(cl.Dim))
	c.suggList.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.NewHexColor(cl.SelectedBg)).
		Foreground(tcell.NewHexColor(cl.Accent)))
	c.suggList.SetHighlightFullLine(true)
	c.suggList.SetBorder(true)
	c.suggList.SetBorderColor(tcell.NewHexColor(cl.Border))
	c.suggList.SetTitle(fmt.Sprintf(" [#%06X]↑↓ navigate · Tab: complete · Esc: close[-] ", cl.Dim))

	c.input = tview.NewInputField()
	c.input.SetLabel(fmt.Sprintf("  [#%06X]›[-] ", cl.Accent))
	c.input.SetLabelColor(tcell.NewHexColor(cl.Label))
	c.input.SetFieldBackgroundColor(tcell.NewHexColor(cl.InputBg))
	c.input.SetFieldTextColor(tcell.NewHexColor(cl.InputText))
	c.input.SetBackgroundColor(tcell.NewHexColor(cl.Background))
	c.input.SetBorder(true)
	c.input.SetBorderColor(tcell.NewHexColor(cl.InputBorder))
	c.input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			c.submitInput()
		case tcell.KeyEscape:
			c.hideSuggestions()
		}
	})

	c.updateHeader()
	c.updateFooter()

	c.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(c.header, 1, 0, false).
		AddItem(c.chatLog, 0, 1, false).
		AddItem(c.statusLine, 1, 0, false).
		AddItem(c.footer, 1, 0, false).
		AddItem(c.input, 3, 0, true)

	c.pages = tview.NewPages()
	c.pages.AddPage("main", c.layout, true, true)

	c.app = tview.NewApplication()
	c.app.SetRoot(c.pages, true).EnableMouse(false)
	c.app.SetFocus(c.input)
}

func (c *chatUI) modeColor() string {
	t := GetTheme()
	switch c.mode {
	case modePlan:
		return fmt.Sprintf("#%06X", t.Colors.ModePlan)
	case modeChat:
		return fmt.Sprintf("#%06X", t.Colors.ModeChat)
	default:
		return fmt.Sprintf("#%06X", t.Colors.ModeBuild)
	}
}

func (c *chatUI) updateHeader() {
	t := GetTheme()
	cl := t.Colors
	modeLabel := fmt.Sprintf("[%s::b]%s[-]", c.modeColor(), c.mode.String())
	c.header.SetText(fmt.Sprintf(
		"  [#%06X::b]HERON[-]  [#%06X]·[-]  [#%06X]%s[-]  [#%06X]·[-]  %s  [#%06X]·[-]  [#%06X]session:%s[-]",
		cl.Accent, cl.Border, cl.Dim, c.modelName, cl.Border, modeLabel, cl.Border, cl.Dim, c.shortSession(),
	))
}

func (c *chatUI) updateFooter() {
	t := GetTheme()
	cl := t.Colors
	c.footer.SetText(fmt.Sprintf(
		"  [#%06X]%s[-]  [#%06X]·[-]  [#%06X]session:%s[-]  [#%06X]·[-]  [#%06X]Tab[-][#%06X]:chat/plan/build  [#%06X]Ctrl+L[-][#%06X]:models  [#%06X]Ctrl+G[-][#%06X]:agents  [#%06X]Ctrl+P[-][#%06X]:sessions  /help[-]",
		cl.Dim, c.modelName, cl.Border, cl.Dim, c.shortSession(), cl.Border,
		cl.Accent, cl.Dim, cl.Accent, cl.Dim, cl.Accent, cl.Dim, cl.Accent, cl.Dim,
	))
}

func (c *chatUI) toggleMode() {
	t := GetTheme()
	switch c.mode {
	case modeChat:
		c.mode = modePlan
	case modePlan:
		c.mode = modeBuild
	default:
		c.mode = modeChat
	}
	c.updateHeader()
	c.setTemporaryStatus(fmt.Sprintf("[#%06X]Switched to %s mode[-]", t.Colors.Accent, c.mode.String()), 2*time.Second)
}

func (c *chatUI) printWelcome() {
	t := GetTheme()
	cl := t.Colors
	fmt.Fprintf(c.chatLog, "\n")
	fmt.Fprintf(c.chatLog, "[#%06X::b]  __                  __               [-]\n", cl.Accent)
	fmt.Fprintf(c.chatLog, "[#%06X::b] / _|  _ __   _   _ | |_   ____  _ __ [-]\n", cl.Accent)
	fmt.Fprintf(c.chatLog, "[#%06X::b]| |_  | '_ \\ | | | || __| |_  / | '_ \\[-]\n", cl.Accent)
	fmt.Fprintf(c.chatLog, "[#%06X::b]|  _| | | | || |_| || |_   / /  | | | |[-]\n", cl.Accent)
	fmt.Fprintf(c.chatLog, "[#%06X::b]|_|   |_| |_| \\__,_| \\__| /___| |_| |_|\n", cl.Accent)
	fmt.Fprintf(c.chatLog, "\n")
	fmt.Fprintf(c.chatLog, "  [#%06X]%-41s[-]\n", cl.Border, strings.Repeat("─", 41))
	fmt.Fprintf(c.chatLog, "  [#%06X]model:[-] [#%06X]%s[-]  [#%06X]· session:[-] [#%06X]%s[-]\n",
		cl.Dim, cl.Accent, c.modelName, cl.Dim, cl.Accent, c.shortSession())
	fmt.Fprintf(c.chatLog, "  [#%06X]type a message to begin · / for commands · Tab to cycle Chat/Plan/Build[-]\n", cl.Dim)
	fmt.Fprintf(c.chatLog, "  [#%06X]%-41s[-]\n\n", cl.Border, strings.Repeat("─", 41))
}

func (c *chatUI) run(ctx context.Context, loop *pkgagent.AgentLoop) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.ctx = ctx

	// Cache installed skill names for shorthand dispatch
	c.skillNames = make(map[string]bool)
	if info := loop.GetStartupInfo(); info != nil {
		if skills, ok := info["skills"].(map[string]any); ok {
			if names, ok := skills["names"].([]string); ok {
				for _, n := range names {
					c.skillNames[strings.ToLower(n)] = true
				}
			}
		}
	}

	// Build suggestion list (static commands + skill shortcuts)
	c.allSugg = []cmdSuggestion{
		{"/help", "show help"},
		{"/clear", "clear chat log"},
		{"/exit", "exit chat"},
		{"/quit", "exit chat"},
		{"/model", "show or switch model"},
		{"/session", "show or change session"},
		{"/skills", "list installed skills"},
		{"/use", "invoke a skill: /use <skill> [msg]"},
		{"/status", "agent status"},
		{"/think", "toggle extended thinking"},
		{"/fast", "toggle fast mode"},
		{"/memory", "search agent memory"},
		{"/list", "list resources (models, skills)"},
		{"/show", "show current settings"},
	}
	skillCmds := make([]cmdSuggestion, 0, len(c.skillNames))
	for name := range c.skillNames {
		skillCmds = append(skillCmds, cmdSuggestion{"/" + name, "skill: " + name})
	}
	sort.Slice(skillCmds, func(i, j int) bool { return skillCmds[i].cmd < skillCmds[j].cmd })
	c.allSugg = append(c.allSugg, skillCmds...)

	sub := loop.SubscribeEvents(64)
	defer loop.UnsubscribeEvents(sub.ID)
	go c.handleEvents(ctx, sub.C)

	go c.runSpinner(ctx)

	// ── Global key capture ────────────────────────────────────
	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyCtrlD:
			cancel()
			c.app.Stop()
			return nil
		case tcell.KeyCtrlL:
			c.app.QueueUpdateDraw(func() { c.showModelPicker() })
			return nil
		case tcell.KeyCtrlG:
			c.app.QueueUpdateDraw(func() { c.showAgentPicker() })
			return nil
		case tcell.KeyCtrlP:
			c.app.QueueUpdateDraw(func() { c.showSessionPicker() })
			return nil
		case tcell.KeyEscape:
			if c.suggVisible {
				c.app.QueueUpdateDraw(c.hideSuggestions)
				return nil
			}
			if c.pages.HasPage("model-picker") {
				c.hideModelPicker()
				return nil
			}
			if c.pages.HasPage("agent-picker") {
				c.hideAgentPicker()
				return nil
			}
			if c.pages.HasPage("session-picker") {
				c.hideSessionPicker()
				return nil
			}
		}
		return event
	})

	// ── Input key capture ─────────────────────────────────────
	c.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if c.suggVisible {
				c.app.QueueUpdateDraw(c.applySugg)
				return nil
			}
			// Tab with empty input → toggle Plan/Build mode
			if strings.TrimSpace(c.input.GetText()) == "" {
				t := GetTheme()
				c.app.QueueUpdateDraw(func() {
					c.toggleMode()
					c.input.SetFieldBackgroundColor(tcell.NewHexColor(t.Colors.SelectedBg))
					time.AfterFunc(200*time.Millisecond, func() {
						c.app.QueueUpdateDraw(func() {
							c.input.SetFieldBackgroundColor(tcell.NewHexColor(t.Colors.InputBg))
						})
					})
				})
				return nil
			}
			// Tab with typed slash prefix → open suggestions
			if strings.HasPrefix(c.input.GetText(), "/") {
				c.app.QueueUpdateDraw(func() { c.updateSuggestions(c.input.GetText()) })
				return nil
			}
			return nil

		case tcell.KeyEscape:
			if c.suggVisible {
				c.app.QueueUpdateDraw(c.hideSuggestions)
				return nil
			}
			return event

		case tcell.KeyUp:
			if c.suggVisible {
				c.app.QueueUpdateDraw(func() { c.moveSugg(-1) })
				return nil
			}
			// History navigation
			c.mu.Lock()
			histIdx := c.histIdx
			c.mu.Unlock()
			if histIdx > 0 {
				newIdx := histIdx - 1
				c.mu.Lock()
				text := c.history[newIdx]
				c.histIdx = newIdx
				c.mu.Unlock()
				c.app.QueueUpdateDraw(func() { c.input.SetText(text) })
			}
			return nil

		case tcell.KeyDown:
			if c.suggVisible {
				c.app.QueueUpdateDraw(func() { c.moveSugg(1) })
				return nil
			}
			// History navigation
			c.mu.Lock()
			histLen := len(c.history)
			histIdx := c.histIdx
			c.mu.Unlock()
			if histIdx < histLen-1 {
				newIdx := histIdx + 1
				c.mu.Lock()
				text := c.history[newIdx]
				c.histIdx = newIdx
				c.mu.Unlock()
				c.app.QueueUpdateDraw(func() { c.input.SetText(text) })
			} else {
				c.mu.Lock()
				c.histIdx = histLen
				c.mu.Unlock()
				c.app.QueueUpdateDraw(func() { c.input.SetText("") })
			}
			return nil
		}
		return event
	})

	// ── Input text change → show/filter suggestions ───────────
	c.input.SetChangedFunc(func(text string) {
		if strings.HasPrefix(text, "/") {
			c.app.QueueUpdateDraw(func() { c.updateSuggestions(text) })
		} else if c.suggVisible {
			c.app.QueueUpdateDraw(c.hideSuggestions)
		}
	})

	c.printWelcome()
	return c.app.Run()
}

// ── Suggestion helpers ────────────────────────────────────────────────────────

func (c *chatUI) updateSuggestions(text string) {
	c.suggList.Clear()
	suggestions := c.suggestionsForInput(text)
	count := 0
	for _, s := range suggestions {
		cmd, desc := s.cmd, s.desc
		c.suggList.AddItem(cmd, desc, 0, func() {
			c.input.SetText(cmd + " ")
			c.hideSuggestions()
		})
		count++
		if count >= 9 {
			break
		}
	}
	if count == 0 {
		c.hideSuggestions()
		return
	}
	c.showSuggestions()
}

func (c *chatUI) hasExactSuggestionMatch(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	for _, s := range c.suggestionsForInput(text) {
		if strings.ToLower(s.cmd) == lower {
			return true
		}
	}
	return false
}

func (c *chatUI) suggestionsForInput(text string) []cmdSuggestion {
	return suggestionsForInput(text, c.allSugg, c.listModelChoices())
}

func (c *chatUI) submitInput() {
	text := strings.TrimSpace(c.input.GetText())
	if c.suggVisible && !c.hasExactSuggestionMatch(text) {
		c.applySugg()
		return
	}
	if text == "" {
		return
	}

	c.input.SetText("")
	c.hideSuggestions()
	if strings.HasPrefix(text, "/") {
		c.handleSlashCommand(text)
		return
	}
	c.sendToLoop(text)
}

func (c *chatUI) showSuggestions() {
	if c.pages.HasPage("suggestions") {
		c.pages.RemovePage("suggestions")
	}
	n := c.suggList.GetItemCount()
	if n == 0 {
		c.suggVisible = false
		return
	}
	h := n*2 + 2
	if h > 22 {
		h = 22
	}
	t := GetTheme()
	overlay := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewBox().SetBackgroundColor(tcell.NewHexColor(t.Colors.Background)), 0, 1, false).
		AddItem(c.suggList, h, 0, false).
		AddItem(tview.NewBox().SetBackgroundColor(tcell.NewHexColor(t.Colors.Background)), 5, 0, false)
	c.pages.AddPage("suggestions", overlay, true, true)
	c.app.SetFocus(c.input)
	c.suggVisible = true
}

func (c *chatUI) hideSuggestions() {
	if c.pages.HasPage("suggestions") {
		c.pages.RemovePage("suggestions")
	}
	c.suggVisible = false
	c.app.SetFocus(c.input)
}

func (c *chatUI) moveSugg(delta int) {
	n := c.suggList.GetItemCount()
	if n == 0 {
		return
	}
	idx := c.suggList.GetCurrentItem() + delta
	if idx < 0 {
		idx = 0
	} else if idx >= n {
		idx = n - 1
	}
	c.suggList.SetCurrentItem(idx)
}

func (c *chatUI) applySugg() {
	if !c.suggVisible || c.suggList.GetItemCount() == 0 {
		return
	}
	main, _ := c.suggList.GetItemText(c.suggList.GetCurrentItem())
	c.input.SetText(main + " ")
	c.hideSuggestions()
}

// ── Slash command handler ─────────────────────────────────────────────────────

func (c *chatUI) handleSlashCommand(raw string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return
	}
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		c.showHelp()
		return
	case "/clear":
		c.chatLog.Clear()
		c.printWelcome()
		return
	case "/exit", "/quit":
		c.app.Stop()
		return
	case "/session":
		if len(parts) > 1 {
			c.sessionKey = parts[1]
			c.updateHeader()
			c.updateFooter()
			c.appendSystemMessage("Session set to: " + c.sessionKey)
		} else {
			c.appendSystemMessage("Current session: " + c.sessionKey)
		}
		return
	case "/model":
		if len(parts) > 1 {
			if err := c.switchModel(strings.Join(parts[1:], " ")); err != nil {
				c.appendError(err.Error())
			}
		} else {
			c.showModelPicker()
		}
		return
	case "/skills":
		raw = "/list skills"
	}

	if skill := strings.TrimPrefix(cmd, "/"); c.skillNames[skill] {
		t := GetTheme()
		c.setTemporaryStatus(fmt.Sprintf("[#%06X]%s[-] [#%06X]executing...[-]", t.Colors.Accent, skill, t.Colors.Dim), 3*time.Second)
		if len(parts) > 1 {
			raw = "/use " + skill + " " + strings.Join(parts[1:], " ")
		} else {
			raw = "/use " + skill
		}
	}

	// Forward to agent loop
	c.sendToLoop(raw)
}

func (c *chatUI) sendToLoop(text string) {
	c.mu.Lock()
	c.history = append(c.history, text)
	c.histIdx = len(c.history)
	c.busy = true
	c.startTime = time.Now()
	c.lastTool = ""
	mode := c.mode
	c.mu.Unlock()

	c.appendUserMessage(text)

	// Apply mode-specific prefix
	payload := text
	switch mode {
	case modePlan:
		payload = "PLAN: " + text
	case modeBuild:
		payload = "BUILD: " + text
	}

	go func() {
		resp, err := c.agentLoop.ProcessDirect(c.ctx, payload, c.sessionKey)
		c.mu.Lock()
		c.busy = false
		c.mu.Unlock()
		t := GetTheme()
		c.app.QueueUpdateDraw(func() {
			if err != nil {
				c.appendError(err.Error())
			} else if resp != "" {
				c.appendAssistantMessage(resp)
			}
			c.statusLine.SetText(fmt.Sprintf("  [#%06X]idle[-]", t.Colors.Dim))
			c.chatLog.ScrollToEnd()
			c.app.SetFocus(c.input)
		})
	}()
}

func (c *chatUI) showHelp() {
	c.appendSystemMessage(
		"TUI commands:\n" +
			"  /help               this message\n" +
			"  /clear              clear chat log\n" +
			"  /session [key]      show or change session\n" +
			"  /model [name]       show or switch model\n" +
			"  /exit  /quit        exit\n" +
			"\n" +
			"Agent commands (forwarded to loop):\n" +
			"  /skills             list installed skills\n" +
			"  /use <skill> [msg]  invoke a skill\n" +
			"  /<skill> [msg]      shorthand for /use\n" +
			"  /status             agent status\n" +
			"  /list models        list models\n" +
			"  /show model         show current model\n" +
			"  /think              toggle extended thinking\n" +
			"  /fast               toggle fast mode\n" +
			"  /memory <query>     search memory\n" +
			"\n" +
			"Keys:\n" +
			"  Tab (empty input)   toggle Chat / Plan / Build mode\n" +
			"  Tab (/ prefix)      open command suggestions\n" +
			"  Tab (suggestions)   complete highlighted suggestion\n" +
			"  Ctrl+L              model picker\n" +
			"  Ctrl+G              agent picker\n" +
			"  Ctrl+P              session picker\n" +
			"  ↑↓                  navigate suggestions / input history\n" +
			"  Esc                 close popup\n" +
			"  Ctrl+C              quit",
	)
}

// ── Model picker ─────────────────────────────────────────────────────────────

func (c *chatUI) showModelPicker() {
	if c.pages.HasPage("model-picker") {
		return
	}

	choices := c.listModelChoices()
	if len(choices) == 0 {
		c.appendError("No models available in the current config")
		return
	}

	t := GetTheme()
	list := tview.NewList()
	list.ShowSecondaryText(true)
	list.SetBackgroundColor(tcell.NewHexColor(t.Colors.Surface))
	list.SetMainTextColor(tcell.NewHexColor(t.Colors.Text))
	list.SetSecondaryTextColor(tcell.NewHexColor(t.Colors.Dim))
	list.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.NewHexColor(t.Colors.SelectedBg)).
		Foreground(tcell.NewHexColor(t.Colors.Accent)))
	list.SetHighlightFullLine(true)
	list.SetBorder(true)
	list.SetBorderColor(tcell.NewHexColor(t.Colors.Accent))
	list.SetTitle(fmt.Sprintf(" [#%06X::b]MODEL PICKER[-]  [#%06X]↑↓ select · Enter switch · Esc close[-] ", t.Colors.Accent, t.Colors.Dim))
	list.SetTitleColor(tcell.NewHexColor(t.Colors.Accent))

	currentIdx := 0
	for idx, choice := range choices {
		name := choice.name
		modelID := choice.modelID
		if name == c.modelName {
			currentIdx = idx
		}
		list.AddItem(name, modelID, 0, func() {
			c.hideModelPicker()
			if err := c.switchModel(name); err != nil {
				c.appendError(err.Error())
			}
		})
	}
	list.SetCurrentItem(currentIdx)
	list.SetDoneFunc(func() {
		c.hideModelPicker()
	})

	overlay := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(list, 30, 0, true).
			AddItem(tview.NewBox(), 0, 1, false), 0, 3, true).
		AddItem(tview.NewBox(), 0, 1, false)

	c.pages.AddPage("model-picker", overlay, true, true)
	c.app.SetFocus(list)
}

func (c *chatUI) hideModelPicker() {
	c.pages.RemovePage("model-picker")
	c.app.SetFocus(c.input)
}

type agentChoice struct {
	name string
	path string
}

func (c *chatUI) listAgents() []agentChoice {
	coordDir := filepath.Join(os.Getenv("HOME"), ".heron", "coordination")
	entries, err := os.ReadDir(coordDir)
	if err != nil {
		return nil
	}
	var agents []agentChoice
	for _, entry := range entries {
		if entry.IsDir() {
			agents = append(agents, agentChoice{
				name: entry.Name(),
				path: filepath.Join(coordDir, entry.Name()),
			})
		}
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].name < agents[j].name
	})
	return agents
}

func (c *chatUI) showAgentPicker() {
	if c.pages.HasPage("agent-picker") {
		return
	}
	t := GetTheme()
	agents := c.listAgents()
	if len(agents) == 0 {
		c.appendSystemMessage("No agents found in ~/.heron/coordination/")
		return
	}

	list := tview.NewList()
	list.ShowSecondaryText(true)
	list.SetBackgroundColor(tcell.NewHexColor(t.Colors.Surface))
	list.SetMainTextColor(tcell.NewHexColor(t.Colors.Text))
	list.SetSecondaryTextColor(tcell.NewHexColor(t.Colors.Dim))
	list.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.NewHexColor(t.Colors.SelectedBg)).
		Foreground(tcell.NewHexColor(t.Colors.Accent)))
	list.SetHighlightFullLine(true)
	list.SetBorder(true)
	list.SetBorderColor(tcell.NewHexColor(t.Colors.Accent))
	list.SetTitle(fmt.Sprintf(" [#%06X::b]AGENT PICKER[-]  [#%06X]↑↓ select · Enter load · Esc close[-] ", t.Colors.Accent, t.Colors.Dim))
	list.SetTitleColor(tcell.NewHexColor(t.Colors.Accent))

	for _, agent := range agents {
		list.AddItem(agent.name, agent.path, 0, func() {
			c.hideAgentPicker()
			c.setTemporaryStatus(fmt.Sprintf("[#%06X]agent: %s[-]  [#%06X]loaded[-]", t.Colors.Accent, agent.name, t.Colors.Dim), 2*time.Second)
		})
	}

	overlay := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(list, 20, 0, true).
			AddItem(tview.NewBox(), 0, 1, false), 0, 3, true).
		AddItem(tview.NewBox(), 0, 1, false)

	c.pages.AddPage("agent-picker", overlay, true, true)
	c.app.SetFocus(list)
}

func (c *chatUI) hideAgentPicker() {
	c.pages.RemovePage("agent-picker")
	c.app.SetFocus(c.input)
}

type sessionChoice struct {
	name     string
	key      string
	messages int
}

func (c *chatUI) listSessions() []sessionChoice {
	memDir := filepath.Join(os.Getenv("HOME"), ".heron", "memory")
	entries, err := os.ReadDir(memDir)
	if err != nil {
		return nil
	}
	var sessions []sessionChoice
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		sessions = append(sessions, sessionChoice{
			name:     name,
			key:      name,
			messages: 0,
		})
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].name > sessions[j].name
	})
	return sessions
}

func (c *chatUI) showSessionPicker() {
	if c.pages.HasPage("session-picker") {
		return
	}
	t := GetTheme()
	sessions := c.listSessions()
	if len(sessions) == 0 {
		c.appendSystemMessage("No sessions found in ~/.heron/memory/")
		return
	}

	list := tview.NewList()
	list.ShowSecondaryText(true)
	list.SetBackgroundColor(tcell.NewHexColor(t.Colors.Surface))
	list.SetMainTextColor(tcell.NewHexColor(t.Colors.Text))
	list.SetSecondaryTextColor(tcell.NewHexColor(t.Colors.Dim))
	list.SetSelectedStyle(tcell.StyleDefault.
		Background(tcell.NewHexColor(t.Colors.SelectedBg)).
		Foreground(tcell.NewHexColor(t.Colors.Accent)))
	list.SetHighlightFullLine(true)
	list.SetBorder(true)
	list.SetBorderColor(tcell.NewHexColor(t.Colors.Accent))
	list.SetTitle(fmt.Sprintf(" [#%06X::b]SESSION PICKER[-]  [#%06X]↑↓ select · Enter load · Esc close[-] ", t.Colors.Accent, t.Colors.Dim))
	list.SetTitleColor(tcell.NewHexColor(t.Colors.Accent))

	currentIdx := 0
	for idx, session := range sessions {
		desc := fmt.Sprintf("%d messages", session.messages)
		if session.key == c.sessionKey {
			currentIdx = idx
		}
		list.AddItem(session.name, desc, 0, func() {
			c.hideSessionPicker()
			c.sessionKey = session.key
			c.updateHeader()
			c.updateFooter()
			c.setTemporaryStatus(fmt.Sprintf("[#%06X]session: %s[-]  [#%06X]loaded[-]", t.Colors.Accent, session.name, t.Colors.Dim), 2*time.Second)
		})
	}
	list.SetCurrentItem(currentIdx)
	list.SetDoneFunc(func() {
		c.hideSessionPicker()
	})

	overlay := tview.NewFlex().
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(list, 20, 0, true).
			AddItem(tview.NewBox(), 0, 1, false), 0, 3, true).
		AddItem(tview.NewBox(), 0, 1, false)

	c.pages.AddPage("session-picker", overlay, true, true)
	c.app.SetFocus(list)
}

func (c *chatUI) hideSessionPicker() {
	c.pages.RemovePage("session-picker")
	c.app.SetFocus(c.input)
}

func (c *chatUI) switchModel(modelName string) error {
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		return fmt.Errorf("model name is required")
	}

	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	resolved, err := resolveModelChoice(modelName, modelChoicesFromConfig(cfg))
	if err != nil {
		return err
	}
	cfg.Agents.Defaults.ModelName = resolved.name

	provider, modelID, err := providers.CreateProvider(cfg)
	if err != nil {
		return fmt.Errorf("switch model %q: %w", resolved.name, err)
	}
	if modelID != "" {
		cfg.Agents.Defaults.ModelName = modelID
	}

	reloadCtx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()
	if err := c.agentLoop.ReloadProviderAndConfig(reloadCtx, provider, cfg); err != nil {
		if stateful, ok := provider.(providers.StatefulProvider); ok {
			stateful.Close()
		}
		return fmt.Errorf("reload model %q: %w", resolved.name, err)
	}

	c.modelName = cfg.Agents.Defaults.ModelName
	c.updateHeader()
	c.updateFooter()
	c.appendSystemMessage("Model switched to: " + c.modelName)
	return nil
}

func (c *chatUI) listModelChoices() []modelChoice {
	if c == nil || c.agentLoop == nil {
		return nil
	}
	return modelChoicesFromConfig(c.agentLoop.GetConfig())
}

func modelChoicesFromConfig(cfg *pkgconfig.Config) []modelChoice {
	if cfg == nil {
		return nil
	}

	choices := make([]modelChoice, 0, len(cfg.ModelList))
	seen := make(map[string]struct{}, len(cfg.ModelList))
	for _, model := range cfg.ModelList {
		if model == nil || model.IsVirtual() {
			continue
		}
		name := strings.TrimSpace(model.ModelName)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		choices = append(choices, modelChoice{
			name:    name,
			modelID: strings.TrimSpace(model.Model),
		})
	}
	sort.Slice(choices, func(i, j int) bool {
		return strings.ToLower(choices[i].name) < strings.ToLower(choices[j].name)
	})
	return choices
}

func suggestionsForInput(text string, base []cmdSuggestion, models []modelChoice) []cmdSuggestion {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(trimmed, "/") {
		return nil
	}

	if lower == "/model" || strings.HasPrefix(lower, "/model ") {
		query := strings.TrimSpace(strings.TrimPrefix(lower, "/model"))
		return modelCommandSuggestions(query, models)
	}

	suggestions := make([]cmdSuggestion, 0, len(base))
	for _, s := range base {
		if lower == "/" || strings.HasPrefix(strings.ToLower(s.cmd), lower) {
			suggestions = append(suggestions, s)
		}
	}
	return suggestions
}

func modelCommandSuggestions(query string, models []modelChoice) []cmdSuggestion {
	if len(models) == 0 {
		return nil
	}

	query = strings.TrimSpace(strings.ToLower(query))
	prefixMatches := make([]cmdSuggestion, 0, len(models))
	containsMatches := make([]cmdSuggestion, 0, len(models))
	for _, choice := range models {
		nameLower := strings.ToLower(choice.name)
		suggestion := cmdSuggestion{
			cmd:  "/model " + choice.name,
			desc: choice.modelID,
		}
		switch {
		case query == "":
			prefixMatches = append(prefixMatches, suggestion)
		case strings.HasPrefix(nameLower, query):
			prefixMatches = append(prefixMatches, suggestion)
		case strings.Contains(nameLower, query):
			containsMatches = append(containsMatches, suggestion)
		}
	}
	return append(prefixMatches, containsMatches...)
}

func resolveModelChoice(input string, models []modelChoice) (modelChoice, error) {
	name := strings.TrimSpace(input)
	if name == "" {
		return modelChoice{}, fmt.Errorf("model name is required")
	}

	lower := strings.ToLower(name)
	var prefixMatches []modelChoice
	for _, choice := range models {
		choiceLower := strings.ToLower(choice.name)
		if choiceLower == lower {
			return choice, nil
		}
		if strings.HasPrefix(choiceLower, lower) {
			prefixMatches = append(prefixMatches, choice)
		}
	}

	switch len(prefixMatches) {
	case 1:
		return prefixMatches[0], nil
	case 0:
		return modelChoice{}, fmt.Errorf("unknown model %q", name)
	default:
		names := make([]string, 0, len(prefixMatches))
		for _, match := range prefixMatches {
			names = append(names, match.name)
		}
		sort.Strings(names)
		return modelChoice{}, fmt.Errorf("model %q is ambiguous: %s", name, strings.Join(names, ", "))
	}
}

// ── Event loop ────────────────────────────────────────────────────────────────

func (c *chatUI) handleEvents(ctx context.Context, ch <-chan pkgagent.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			switch evt.Kind {
			case pkgagent.EventKindToolExecStart:
				if p, ok := evt.Payload.(pkgagent.ToolExecStartPayload); ok {
					var argParts []string
					for k, v := range p.Arguments {
						argParts = append(argParts, fmt.Sprintf("%s=%v", k, v))
					}
					argsStr := strings.Join(argParts, ", ")
					c.mu.Lock()
					c.lastTool = p.Tool
					c.mu.Unlock()
					tool, args := p.Tool, argsStr
					c.app.QueueUpdateDraw(func() { c.appendToolStart(tool, args) })
				}
			case pkgagent.EventKindToolExecEnd:
				if p, ok := evt.Payload.(pkgagent.ToolExecEndPayload); ok {
					isErr, dur, tool := p.IsError, p.Duration, p.Tool
					c.app.QueueUpdateDraw(func() { c.appendToolEnd(tool, dur, isErr) })
				}
			case pkgagent.EventKindLLMRetry:
				if p, ok := evt.Payload.(pkgagent.LLMRetryPayload); ok {
					reason, attempt := p.Reason, p.Attempt
					t := GetTheme()
					c.app.QueueUpdateDraw(func() {
						fmt.Fprintf(c.chatLog, "  [#%06X]↺  retrying: %s (attempt %d)[-]\n",
							t.Colors.Error, tview.Escape(reason), attempt)
					})
				}
			case pkgagent.EventKindError:
				if p, ok := evt.Payload.(pkgagent.ErrorPayload); ok {
					msg, stage := p.Message, p.Stage
					c.app.QueueUpdateDraw(func() {
						c.appendError(fmt.Sprintf("[%s] %s", stage, msg))
					})
				}
			}
		}
	}
}

func (c *chatUI) setTemporaryStatus(text string, duration time.Duration) {
	c.mu.Lock()
	c.tempStatus = text
	c.mu.Unlock()

	c.app.QueueUpdateDraw(func() {
		c.statusLine.SetText("  " + text)
	})

	go func() {
		time.Sleep(duration)
		c.mu.Lock()
		c.tempStatus = ""
		c.mu.Unlock()
	}()
}

func (c *chatUI) runSpinner(ctx context.Context) {
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()
	phraseIdx := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.mu.Lock()
			busy := c.busy
			tool := c.lastTool
			tempStatus := c.tempStatus
			elapsed := time.Since(c.startTime)
			frame := spinnerFrames[c.spinIdx%len(spinnerFrames)]
			if busy {
				c.spinIdx++
			}
			c.mu.Unlock()

			if tempStatus != "" {
				c.app.QueueUpdateDraw(func() { c.statusLine.SetText("  " + tempStatus) })
			} else if busy {
				t := GetTheme()
				phrase := waitingPhrases[phraseIdx%len(waitingPhrases)]
				if c.spinIdx%8 == 0 {
					phraseIdx++
				}
				var status string
				if tool != "" {
					status = fmt.Sprintf("  [#%06X]%s[-]  [#%06X]%s · %.1fs[-]", t.Colors.Spinner, frame, t.Colors.Dim, tool, elapsed.Seconds())
				} else {
					status = fmt.Sprintf("  [#%06X]%s[-]  [#%06X]%s · %.1fs[-]", t.Colors.Spinner, frame, t.Colors.Dim, phrase, elapsed.Seconds())
				}
				c.app.QueueUpdateDraw(func() { c.statusLine.SetText(status) })
			} else {
				t := GetTheme()
				c.app.QueueUpdateDraw(func() { c.statusLine.SetText(fmt.Sprintf("  [#%06X]idle[-]", t.Colors.Dim)) })
			}
		}
	}
}

// ── Message rendering ─────────────────────────────────────────────────────────

func (c *chatUI) appendUserMessage(text string) {
	t := GetTheme()
	cl := t.Colors
	fmt.Fprintf(c.chatLog, "\n[#%06X:#%06X:b] You [-:-:-]\n", cl.Accent, cl.SelectedBg)
	for _, line := range strings.Split(text, "\n") {
		fmt.Fprintf(c.chatLog, "[#%06X:#%06X:-] %s [-:-:-]\n", cl.UserText, cl.SelectedBg, tview.Escape(line))
	}
	fmt.Fprintf(c.chatLog, "\n")
	c.chatLog.ScrollToEnd()
}

func (c *chatUI) appendAssistantMessage(text string) {
	t := GetTheme()
	cl := t.Colors
	fmt.Fprintf(c.chatLog, "[#%06X]●[-]  [#%06X::b]Heron[-]\n\n", cl.Accent, cl.Accent)
	rendered := c.renderMarkdown(text)
	fmt.Fprintf(c.chatLog, "%s\n\n", rendered)
}

func (c *chatUI) appendSystemMessage(text string) {
	t := GetTheme()
	cl := t.Colors
	lines := strings.Split(text, "\n")
	fmt.Fprintf(c.chatLog, "\n")
	for _, line := range lines {
		fmt.Fprintf(c.chatLog, "  [#%06X]%s[-]\n", cl.SystemText, line)
	}
	fmt.Fprintf(c.chatLog, "\n")
	c.chatLog.ScrollToEnd()
}

func (c *chatUI) appendToolStart(tool, args string) {
	t := GetTheme()
	cl := t.Colors
	const inner = cardWidth - 2
	if len(tool) > inner-6 {
		tool = tool[:inner-9] + "..."
	}
	titlePart := "─ " + tool + " "
	dashes := inner - 1 - len(titlePart)
	if dashes < 0 {
		dashes = 0
	}
	fmt.Fprintf(c.chatLog, "  [#%06X]┌[#%06X::i]%s[-][#%06X]%s┐[-]\n",
		cl.Border, cl.Dim, titlePart, cl.Border, strings.Repeat("─", dashes))
	fmt.Fprintf(c.chatLog, "  [#%06X]│[#%06X] ⟳  %s%s[#%06X]│[-]\n",
		cl.Border, cl.Dim, tview.Escape(args), strings.Repeat(" ", inner-5-len(args)), cl.Border)
}

func (c *chatUI) appendToolEnd(_ string, dur time.Duration, isErr bool) {
	t := GetTheme()
	cl := t.Colors
	const inner = cardWidth - 2
	durStr := fmt.Sprintf("%.2fs", dur.Seconds())
	durPad := inner - 5 - len(durStr)
	if durPad < 0 {
		durPad = 0
	}
	if isErr {
		fmt.Fprintf(c.chatLog, "  [#%06X]│[-][#%06X] ✗  %s%s[#%06X]│[-]\n",
			cl.Border, cl.Error, durStr, strings.Repeat(" ", durPad), cl.Border)
		fmt.Fprintf(c.chatLog, "  [#%06X]└%s┘[-]\n\n", cl.Border, strings.Repeat("─", cardWidth-2))
	} else {
		fmt.Fprintf(c.chatLog, "  [#%06X]│[-][#%06X] ✓[-][#%06X]  %s%s[#%06X]│[-]\n",
			cl.Border, cl.Success, cl.Dim, durStr, strings.Repeat(" ", durPad), cl.Border)
		fmt.Fprintf(c.chatLog, "  [#%06X]└%s┘[-]\n\n", cl.Border, strings.Repeat("─", cardWidth-2))
	}
}

func (c *chatUI) appendError(msg string) {
	t := GetTheme()
	cl := t.Colors
	fmt.Fprintf(c.chatLog, "\n  [#%06X]✗  %s[-]\n\n", cl.Error, tview.Escape(msg))
	c.chatLog.ScrollToEnd()
}

// ── Lightweight Markdown Renderer ─────────────────────────────────────────────

func (c *chatUI) renderMarkdown(text string) string {
	t := GetTheme()
	cl := t.Colors
	lines := strings.Split(text, "\n")
	var out strings.Builder

	codeBlock := false
	codeLang := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if codeBlock {
				out.WriteString(fmt.Sprintf("  [#%06X]┌%s┐[-]\n", cl.Border, strings.Repeat("─", cardWidth-4)))
				out.WriteString(fmt.Sprintf("  [#%06X]│[-]  [#%06X]%s[-]  [#%06X]│[-]\n",
					cl.Border, cl.CodeBlock, strings.Repeat(" ", cardWidth-6), cl.Border))
				out.WriteString(fmt.Sprintf("  [#%06X]└%s┘[-]\n", cl.Border, strings.Repeat("─", cardWidth-4)))
				codeBlock = false
				codeLang = ""
			} else {
				codeBlock = true
				codeLang = strings.TrimPrefix(line, "```")
				out.WriteString(fmt.Sprintf("  [#%06X]┌%s┐[-]\n", cl.Border, strings.Repeat("─", cardWidth-4)))
				out.WriteString(fmt.Sprintf("  [#%06X]│[-]  [#%06X]%s[-]  [#%06X]│[-]\n",
					cl.Border, cl.Border, codeLang, cl.Border))
			}
			continue
		}

		if codeBlock {
			escaped := tview.Escape(line)
			if len(escaped) > cardWidth-6 {
				escaped = escaped[:cardWidth-9] + "..."
			}
			out.WriteString(fmt.Sprintf("  [#%06X]│[-]  [#%06X]%-*s[-]  [#%06X]│[-]\n",
				cl.Border, cl.CodeBlock, cardWidth-6, escaped, cl.Border))
			continue
		}

		if strings.HasPrefix(strings.TrimSpace(line), "> ") {
			quote := strings.TrimSpace(strings.TrimPrefix(line, "> "))
			escaped := tview.Escape(quote)
			out.WriteString(fmt.Sprintf("  [#%06X]│[-][#%06X] %s [-:-:-][#%06X]│[-]\n",
				cl.Border, cl.QuoteBorder, escaped, cl.Border))
			continue
		}

		rendered := c.renderInline(line)
		out.WriteString(fmt.Sprintf("  %s\n", rendered))
	}

	return out.String()
}

func (c *chatUI) renderInline(line string) string {
	t := GetTheme()
	cl := t.Colors

	line = tview.Escape(line)

	headerRegex := regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	if matches := headerRegex.FindStringSubmatch(line); matches != nil {
		level := len(matches[1])
		text := matches[2]
		boldness := strings.Repeat(":", 7-level) + "b"
		return fmt.Sprintf("[#%06X::%s]%s[-]", cl.Accent, boldness, text)
	}

	line = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllStringFunc(line, func(m string) string {
		content := m[2 : len(m)-2]
		return fmt.Sprintf("[#%06X::b]%s[-]", cl.Text, content)
	})

	line = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllStringFunc(line, func(m string) string {
		content := m[1 : len(m)-1]
		return fmt.Sprintf("[#%06X::i]%s[-]", cl.Text, content)
	})

	line = regexp.MustCompile("`([^`]+)`").ReplaceAllStringFunc(line, func(m string) string {
		content := m[1 : len(m)-1]
		return fmt.Sprintf("[#%06X]%s[-]", cl.Code, content)
	})

	line = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllStringFunc(line, func(m string) string {
		parts := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).FindStringSubmatch(m)
		if len(parts) >= 2 {
			return fmt.Sprintf("[#%06X]%s[-] ([#%06X]link[-])", cl.Link, parts[1], cl.Dim)
		}
		return m
	})

	return fmt.Sprintf("[#%06X]%s[-]", cl.Text, line)
}
