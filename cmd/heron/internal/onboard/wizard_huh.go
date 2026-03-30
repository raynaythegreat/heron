package onboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/raynaythegreat/heron/cmd/heron/internal"
	authpkg "github.com/raynaythegreat/heron/cmd/heron/internal/auth"
	"github.com/raynaythegreat/heron/pkg/config"
)

const totalSteps = 8

func runHuhWizard(encrypt bool) {
	configPath := internal.GetConfigPath()

	fmt.Println()
	fmt.Println(renderWelcomeScreen())
	fmt.Println()

	configExists := false
	if _, err := os.Stat(configPath); err == nil {
		configExists = true
		var fresh bool
		freshForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Existing configuration found").
					Description(fmt.Sprintf("Config file: %s\n\nStart fresh? This will reset your configuration.", configPath)).
					Next(true),
				huh.NewConfirm().
					Title("Start fresh?").
					Description("This will reset your configuration to defaults.").
					Affirmative("Yes, start fresh").
					Negative("No, keep existing").
					Value(&fresh),
			),
		).WithTheme(heronTheme())
		_ = freshForm.Run()
		if !fresh {
			configExists = true
		} else {
			configExists = false
		}
	}

	var cfg *config.Config
	if configExists {
		var err error
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = config.DefaultConfig()
	}

	if encrypt {
		passphrase, err := promptPassphrase()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		os.Setenv("HERON_KEY_PASSPHRASE", passphrase)
		if err = setupSSHKey(); err != nil {
			fmt.Printf("Error generating SSH key: %v\n", err)
			os.Exit(1)
		}
	}

	var (
		selectedIdxs []int
		configured   []string
		primary      string
	)

	selectedIdxs = step1SelectProviders()

	enterAPIKeysHuh(cfg, selectedIdxs)

	if reloaded, err := config.LoadConfig(configPath); err == nil {
		cfg = reloaded
	}
	configured = configuredModels(cfg, selectedIdxs)

	if len(configured) == 0 {
		fmt.Println()
		fmt.Println(formatWarning("No providers configured — skipping model selection."))
		fmt.Println(formatDim("You can add providers later by editing ~/.heron/config.json"))
	} else {
		primary = step3PrimaryModel(configured)
		cfg.Agents.Defaults.ModelName = primary

		step4FallbackModels(cfg, configured, primary)

		step5SmartRouting(cfg, configured, primary)
	}

	step6Tools(cfg)

	step7Ollama(cfg)

	step8MCPCatalog(cfg, selectedIdxs)

	pruneUnconfigured(cfg, selectedIdxs)

	createWorkspaceTemplates(cfg.WorkspacePath())
	extractBuiltinSkills()

	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	syncToTUIConfig(cfg, primary)

	fmt.Println()
	fmt.Println(renderCompletionSummary(cfg, configPath))
	fmt.Println()
}

func step1SelectProviders() []int {
	type providerKey struct {
		idx int
	}

	options := make([]huh.Option[providerKey], len(providerCatalog))
	for i, p := range providerCatalog {
		tag := ""
		if !p.NeedsKey {
			tag = " [free]"
		}
		if len(p.AuthMethods) > 0 {
			tag += " [OAuth]"
		}
		options[i] = huh.NewOption(
			fmt.Sprintf("%s%s  — %s", p.Name, tag, p.Desc),
			providerKey{idx: i},
		)
	}

	var selected []providerKey
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[providerKey]().
				Title(stepHeader(1, totalSteps, "Select Providers")).
				Description("Choose the AI providers you want to configure.\n↑/↓ navigate  ·  Space toggle  ·  / filter  ·  Enter confirm").
				Options(options...).
				Filtering(true).
				Value(&selected),
		),
	).WithTheme(heronTheme())

	if err := form.Run(); err != nil {
		return nil
	}

	var idxs []int
	for _, s := range selected {
		if s.idx >= 0 {
			idxs = append(idxs, s.idx)
		}
	}
	return idxs
}

func enterAPIKeysHuh(cfg *config.Config, selectedIdxs []int) {
	for _, idx := range selectedIdxs {
		if idx >= len(providerCatalog) {
			continue
		}
		p := providerCatalog[idx]

		if !p.NeedsKey {
			continue
		}

		if len(p.AuthMethods) > 0 {
			type authChoice struct {
				index int
			}
			var chosen authChoice
			authOptions := make([]huh.Option[authChoice], len(p.AuthMethods))
			for i, m := range p.AuthMethods {
				authOptions[i] = huh.NewOption(m.Label, authChoice{index: i})
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[authChoice]().
						Title(stepHeader(2, totalSteps, fmt.Sprintf("Authenticate — %s", p.Name))).
						Description("Choose an authentication method.").
						Options(authOptions...).
						Value(&chosen),
				),
			).WithTheme(heronTheme())

			if err := form.Run(); err != nil {
				continue
			}

			method := p.AuthMethods[chosen.index]
			if method.IsOAuth {
				var done bool
				_ = spinner.New().Title(fmt.Sprintf("  Authenticating with %s via OAuth...", p.Name)).Action(func() {
					if err := authpkg.LoginProvider(method.OAuthArg, method.DevCode, true, method.BrowserOAuth); err != nil {
						fmt.Println(formatError(fmt.Sprintf("OAuth failed: %v", err)))
					} else {
						done = true
					}
				}).Run()

				if done {
					fmt.Println(formatSuccess(fmt.Sprintf("%s authenticated", p.Name)))
				} else {
					fmt.Println(formatWarning(fmt.Sprintf("Skipped %s", p.Name)))
				}
				continue
			}
		}

		var apiKey string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(stepHeader(2, totalSteps, fmt.Sprintf("API Key — %s", p.Name))).
					Description("Paste your API key below (input is hidden).").
					Placeholder("sk-...").
					Value(&apiKey).
					Password(true),
			),
		).WithTheme(heronTheme())

		if err := form.Run(); err != nil || strings.TrimSpace(apiKey) == "" {
			fmt.Println(formatWarning(fmt.Sprintf("Skipped %s", p.Name)))
			continue
		}

		key := strings.TrimSpace(apiKey)
		for _, modelName := range p.Models {
			mc, _ := cfg.GetModelConfig(modelName)
			if mc != nil {
				mc.SetAPIKey(key)
			}
		}
		fmt.Println(formatSuccess(fmt.Sprintf("%s API key set", p.Name)))
	}
}

func step3PrimaryModel(configured []string) string {
	var primary string
	options := make([]huh.Option[string], len(configured))
	for i, m := range configured {
		options[i] = huh.NewOption(m, m)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(stepHeader(3, totalSteps, "Primary Model")).
				Description("Which model should agents use by default?").
				Options(options...).
				Filtering(true).
				Value(&primary),
		),
	).WithTheme(heronTheme())

	_ = form.Run()
	return primary
}

func step4FallbackModels(cfg *config.Config, configured []string, primary string) {
	remaining := without(configured, primary)
	if len(remaining) == 0 {
		fmt.Println(formatDim("  No additional models available for fallback."))
		return
	}

	options := make([]huh.Option[string], len(remaining))
	for i, m := range remaining {
		options[i] = huh.NewOption(m, m)
	}

	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(stepHeader(4, totalSteps, "Fallback Models")).
				Description(fmt.Sprintf("When %s hits a rate limit, failover to these (in order).", primary)).
				Options(options...).
				Filtering(true).
				Value(&selected),
		),
	).WithTheme(heronTheme())

	_ = form.Run()
	cfg.Agents.Defaults.ModelFallbacks = selected
}

func step5SmartRouting(cfg *config.Config, configured []string, primary string) {
	var enableRouting bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(stepHeader(5, totalSteps, "Smart Routing")).
				Description("Smart routing sends simple tasks to a cheaper/faster model\nand routes complex tasks (code, long context) to your primary model."),
			huh.NewConfirm().
				Title("Enable smart routing?").
				Affirmative("Yes, enable").
				Negative("No, skip").
				Value(&enableRouting),
		),
	).WithTheme(heronTheme())

	_ = form.Run()

	if !enableRouting {
		return
	}

	lightModel := primary
	remaining := without(configured, primary)
	if len(remaining) > 0 {
		options := make([]huh.Option[string], len(remaining))
		for i, m := range remaining {
			options[i] = huh.NewOption(m, m)
		}

		var chosen string
		lightForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select light model for simple tasks").
					Options(options...).
					Filtering(true).
					Value(&chosen),
			),
		).WithTheme(heronTheme())
		_ = lightForm.Run()
		if chosen != "" {
			lightModel = chosen
		}
	}

	cfg.Agents.Defaults.Routing = &config.RoutingConfig{
		Enabled:    true,
		LightModel: lightModel,
		Threshold:  0.35,
	}
	fmt.Println(formatSuccess(fmt.Sprintf("Routing enabled — simple → %s, complex → %s", lightModel, primary)))
}

func step6Tools(cfg *config.Config) {
	var webEnabled, execEnabled, editEnabled, mcpEnabled bool = true, true, true, true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(stepHeader(6, totalSteps, "Tools")).
				Description("Toggle the tools available to agents."),
			huh.NewConfirm().
				Title("Web Search").
				Description("DuckDuckGo + Brave + Tavily").
				Affirmative("Enabled").
				Negative("Disabled").
				Value(&webEnabled),
			huh.NewConfirm().
				Title("Code / Shell Execution").
				Description("Run commands and scripts").
				Affirmative("Enabled").
				Negative("Disabled").
				Value(&execEnabled),
			huh.NewConfirm().
				Title("File Operations").
				Description("Read, write, and edit files").
				Affirmative("Enabled").
				Negative("Disabled").
				Value(&editEnabled),
			huh.NewConfirm().
				Title("MCP Servers").
				Description("Model Context Protocol integrations").
				Affirmative("Enabled").
				Negative("Disabled").
				Value(&mcpEnabled),
		),
	).WithTheme(heronTheme())

	_ = form.Run()

	cfg.Tools.Web.Enabled = webEnabled
	cfg.Tools.Exec.Enabled = execEnabled
	cfg.Tools.EditFile.Enabled = editEnabled
	cfg.Tools.MCP.Enabled = mcpEnabled
}

func step7Ollama(cfg *config.Config) {
	var ollamaStatus string
	var ollamaModels []string

	_ = spinner.New().Title("  Scanning for Ollama...").Action(func() {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://localhost:11434/api/tags")
		if err != nil {
			ollamaStatus = "not_detected"
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ollamaStatus = "error"
			return
		}

		var tags ollamaTagsResponse
		if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
			ollamaStatus = "error"
			return
		}

		if len(tags.Models) == 0 {
			ollamaStatus = "no_models"
			return
		}

		ollamaStatus = "detected"
		for _, m := range tags.Models {
			ollamaModels = append(ollamaModels, m.Name)
		}
	}).Run()

	fmt.Println()
	switch ollamaStatus {
	case "not_detected":
		fmt.Println(formatDim("  Ollama not detected on localhost:11434"))
		fmt.Println(formatDim("  Install from https://ollama.com"))
	case "error":
		fmt.Println(formatDim("  Ollama detected but returned an error"))
	case "no_models":
		fmt.Println(formatSuccess("Ollama running — no models pulled yet"))
		fmt.Println(formatDim("  Pull models with: ollama pull llama3"))
	case "detected":
		fmt.Println(formatSuccess(fmt.Sprintf("Ollama found — %d model(s):", len(ollamaModels))))
		for _, m := range ollamaModels {
			fmt.Printf("    %s %s\n",
				lipgloss.NewStyle().Foreground(colorMauve).Render("●"),
				lipgloss.NewStyle().Foreground(colorFg).Render(m),
			)
		}
	}
}

type mcpCatalogEntry struct {
	name        string
	description string
	command     string
	args        []string
	envFile     string
	providers   []string
}

var mcpCatalog = []mcpCatalogEntry{
	{
		name:        "filesystem",
		description: "Read/write files in your workspace via MCP",
		command:     "npx",
		args:        []string{"-y", "@anthropic/mcp-server-filesystem", "<workspace>"},
		providers:   []string{"OpenRouter", "Anthropic", "OpenAI", "Google Gemini", "DeepSeek", "Groq", "Mistral AI", "xAI Grok"},
	},
	{
		name:        "sequential-thinking",
		description: "Step-by-step reasoning for complex problems",
		command:     "npx",
		args:        []string{"-y", "@anthropic/mcp-server-sequential-thinking"},
		providers:   []string{"Anthropic", "OpenAI", "Google Gemini"},
	},
	{
		name:        "memory",
		description: "Persistent memory across conversations",
		command:     "npx",
		args:        []string{"-y", "@anthropic/mcp-server-memory"},
		providers:   []string{"Anthropic", "OpenAI"},
	},
	{
		name:        "fetch",
		description: "Fetch and parse web pages",
		command:     "npx",
		args:        []string{"-y", "@anthropic/mcp-server-fetch"},
		providers:   []string{"OpenRouter", "Anthropic", "OpenAI", "Google Gemini", "DeepSeek", "Groq", "Mistral AI", "xAI Grok"},
	},
	{
		name:        "puppeteer",
		description: "Browser automation (screenshots, form filling)",
		command:     "npx",
		args:        []string{"-y", "@anthropic/mcp-server-puppeteer"},
		providers:   []string{"Anthropic", "OpenAI", "Google Gemini"},
	},
	{
		name:        "github",
		description: "GitHub API integration (PRs, issues, repos)",
		command:     "npx",
		args:        []string{"-y", "@anthropic/mcp-server-github"},
		envFile:     ".env.github",
		providers:   []string{"OpenRouter", "Anthropic", "OpenAI", "Google Gemini", "DeepSeek", "Groq", "Mistral AI", "xAI Grok"},
	},
	{
		name:        "brave-search",
		description: "Web search via Brave Search API",
		command:     "npx",
		args:        []string{"-y", "@anthropic/mcp-server-brave-search"},
		envFile:     ".env.brave",
		providers:   []string{"OpenRouter", "Anthropic", "OpenAI", "Google Gemini", "DeepSeek", "Groq", "Mistral AI", "xAI Grok"},
	},
	{
		name:        "notebooklm",
		description: "Generate NotebookLM podcast-style briefings",
		command:     "uvx",
		args:        []string{"--from", "notebooklm-mcp-cli", "notebooklm-mcp"},
		providers:   []string{"Anthropic", "OpenAI", "Google Gemini"},
	},
}

func step8MCPCatalog(cfg *config.Config, selectedIdxs []int) {
	selectedProviders := map[string]bool{}
	for _, idx := range selectedIdxs {
		if idx < len(providerCatalog) {
			selectedProviders[providerCatalog[idx].Name] = true
		}
	}

	alreadyEnabled := map[string]bool{}
	if cfg.Tools.MCP.Servers != nil {
		for name, serverCfg := range cfg.Tools.MCP.Servers {
			if serverCfg.Enabled {
				alreadyEnabled[name] = true
			}
		}
	}

	var relevant []mcpCatalogEntry
	for _, entry := range mcpCatalog {
		for _, p := range entry.providers {
			if selectedProviders[p] {
				relevant = append(relevant, entry)
				break
			}
		}
	}

	fmt.Println()
	fmt.Println(stepHeader(8, totalSteps, "MCP Server Catalog"))
	fmt.Println()

	if len(relevant) == 0 {
		fmt.Println(formatDim("  No MCP servers available for your selected providers."))
		return
	}

	fmt.Println(formatBody("  Recommended MCP servers for your providers:"))
	fmt.Println()

	for _, entry := range relevant {
		status := lipgloss.NewStyle().Foreground(colorSurface2).Render("  ○  available")
		if alreadyEnabled[entry.name] {
			status = lipgloss.NewStyle().Foreground(colorGreen).Render("  ✔  enabled")
		}
		fmt.Printf("    %-22s %s\n", lipgloss.NewStyle().Foreground(colorFg).Bold(true).Render(entry.name), status)
		fmt.Printf("      %s\n", lipgloss.NewStyle().Foreground(colorOverlay0).Render(entry.description))
		cmdLine := entry.command + " " + strings.Join(entry.args, " ")
		fmt.Printf("      %s\n\n", lipgloss.NewStyle().Foreground(colorDim).Render(cmdLine))
	}

	type mcpKey struct {
		index int
	}
	options := make([]huh.Option[mcpKey], 0, len(relevant))
	for i, entry := range relevant {
		prefix := "  "
		if alreadyEnabled[entry.name] {
			prefix = "✔ "
		}
		label := fmt.Sprintf("%s%-22s — %s", prefix, entry.name, entry.description)
		options = append(options, huh.NewOption(label, mcpKey{index: i}))
	}

	var selected []mcpKey
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[mcpKey]().
				Title("Enable MCP Servers").
				Description("Select servers to enable. They auto-install on first use via npx/uvx.").
				Options(options...).
				Filtering(true).
				Value(&selected),
		),
	).WithTheme(heronTheme())

	if err := form.Run(); err != nil {
		return
	}

	if cfg.Tools.MCP.Servers == nil {
		cfg.Tools.MCP.Servers = make(map[string]config.MCPServerConfig)
	}

	enabledCount := 0
	for _, s := range selected {
		if s.index < len(relevant) {
			entry := relevant[s.index]
			if cfg.Tools.MCP.Servers[entry.name].Command == "" {
				serverCfg := config.MCPServerConfig{
					Enabled:  true,
					Command:  entry.command,
					Args:     entry.args,
					Deferred: boolPtr(true),
				}
				if entry.envFile != "" {
					serverCfg.EnvFile = entry.envFile
				}
				cfg.Tools.MCP.Servers[entry.name] = serverCfg
			} else {
				existing := cfg.Tools.MCP.Servers[entry.name]
				existing.Enabled = true
				cfg.Tools.MCP.Servers[entry.name] = existing
			}
			enabledCount++
		}
	}

	if enabledCount > 0 {
		fmt.Println()
		fmt.Println(formatSuccess(fmt.Sprintf("%d MCP server(s) enabled — they will auto-install on first use", enabledCount)))
		if !cfg.Tools.MCP.Enabled {
			cfg.Tools.MCP.Enabled = true
		}
	}
}

func boolPtr(v bool) *bool {
	return &v
}
