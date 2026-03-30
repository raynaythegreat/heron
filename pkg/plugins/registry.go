package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PluginTier string

const (
	TierSystem       PluginTier = "system"
	TierCurated      PluginTier = "curated"
	TierExperimental PluginTier = "experimental"
	TierLocal        PluginTier = "local"
)

type PluginType string

const (
	TypeSkill  PluginType = "skill"
	TypePlugin PluginType = "plugin"
)

type PluginMetadata struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Version      string             `json:"version"`
	Tier         PluginTier         `json:"tier"`
	Type         PluginType         `json:"type"`
	Author       string             `json:"author,omitempty"`
	Tags         []string           `json:"tags,omitempty"`
	Dependencies []string           `json:"dependencies,omitempty"`
	AllowedTools []string           `json:"allowed_tools,omitempty"`
	Requires     PluginRequirements `json:"requires,omitempty"`
	InstallDate  time.Time          `json:"install_date,omitempty"`
	Enabled      bool               `json:"enabled"`
	TrustScore   float64            `json:"trust_score,omitempty"`
	SecurityScan *SecurityReport    `json:"security_scan,omitempty"`
}

type PluginRequirements struct {
	Env  []string `json:"env,omitempty"`
	Bins []string `json:"bins,omitempty"`
}

type SecurityReport struct {
	ScannedAt       time.Time `json:"scanned_at"`
	PromptInjection bool      `json:"prompt_injection"`
	MaliciousCode   bool      `json:"malicious_code"`
	DangerousCmds   bool      `json:"dangerous_cmds"`
	Score           int       `json:"score"`
	Issues          []string  `json:"issues,omitempty"`
}

type PluginRegistry struct {
	pluginsPath string
	plugins     map[string]*PluginMetadata
}

func NewPluginRegistry() *PluginRegistry {
	home, _ := os.UserHomeDir()
	pluginsPath := filepath.Join(home, ".heron", "plugins")
	os.MkdirAll(pluginsPath, 0755)
	os.MkdirAll(filepath.Join(pluginsPath, "system"), 0755)
	os.MkdirAll(filepath.Join(pluginsPath, "curated"), 0755)
	os.MkdirAll(filepath.Join(pluginsPath, "experimental"), 0755)
	os.MkdirAll(filepath.Join(pluginsPath, "local"), 0755)

	return &PluginRegistry{
		pluginsPath: pluginsPath,
		plugins:     make(map[string]*PluginMetadata),
	}
}

func (r *PluginRegistry) LoadAll() error {
	r.plugins = make(map[string]*PluginMetadata)
	tiers := []PluginTier{TierSystem, TierCurated, TierExperimental, TierLocal}
	for _, tier := range tiers {
		tierPath := filepath.Join(r.pluginsPath, string(tier))
		entries, err := os.ReadDir(tierPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				metaPath := filepath.Join(tierPath, entry.Name(), "plugin.json")
				if _, err := os.Stat(metaPath); err != nil {
					skillPath := filepath.Join(tierPath, entry.Name(), "SKILL.md")
					if _, err := os.Stat(skillPath); err != nil {
						continue
					}
					r.plugins[entry.Name()] = &PluginMetadata{
						ID:         entry.Name(),
						Name:       entry.Name(),
						Tier:       tier,
						Type:       TypeSkill,
						Enabled:    true,
						TrustScore: 1.0,
					}
					continue
				}
			}
			if !entry.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(tierPath, entry.Name(), "plugin.json"))
			if err != nil {
				continue
			}
			var meta PluginMetadata
			if err := json.Unmarshal(data, &meta); err != nil {
				continue
			}
			meta.Tier = tier
			r.plugins[meta.ID] = &meta
		}
	}
	return nil
}

func (r *PluginRegistry) List(tier PluginTier) []*PluginMetadata {
	var results []*PluginMetadata
	for _, p := range r.plugins {
		if tier == "" || p.Tier == tier {
			results = append(results, p)
		}
	}
	return results
}

func (r *PluginRegistry) Get(id string) (*PluginMetadata, bool) {
	p, ok := r.plugins[id]
	return p, ok
}

func (r *PluginRegistry) Install(source string, tier PluginTier) (*PluginMetadata, error) {
	id := filepath.Base(source)
	destDir := filepath.Join(r.pluginsPath, string(tier), id)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	pluginMeta := &PluginMetadata{
		ID:          id,
		Name:        id,
		Tier:        tier,
		Type:        TypeLocal,
		InstallDate: time.Now(),
		Enabled:     true,
		TrustScore:  0.5,
	}

	data, _ := json.MarshalIndent(pluginMeta, "", "  ")
	if err := os.WriteFile(filepath.Join(destDir, "plugin.json"), data, 0644); err != nil {
		return nil, err
	}

	r.plugins[id] = pluginMeta
	return pluginMeta, nil
}

func (r *PluginRegistry) Uninstall(id string) error {
	plugin, ok := r.plugins[id]
	if !ok {
		return fmt.Errorf("plugin not found: %s", id)
	}

	pluginDir := filepath.Join(r.pluginsPath, string(plugin.Tier), id)
	if err := os.RemoveAll(pluginDir); err != nil {
		return err
	}

	delete(r.plugins, id)
	return nil
}

func (r *PluginRegistry) Toggle(id string, enabled bool) error {
	plugin, ok := r.plugins[id]
	if !ok {
		return fmt.Errorf("plugin not found: %s", id)
	}
	plugin.Enabled = enabled

	destDir := filepath.Join(r.pluginsPath, string(plugin.Tier), id)
	data, _ := json.MarshalIndent(plugin, "", "  ")
	return os.WriteFile(filepath.Join(destDir, "plugin.json"), data, 0644)
}

func (r *PluginRegistry) ScanSecurity(id string) (*SecurityReport, error) {
	plugin, ok := r.plugins[id]
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", id)
	}

	pluginDir := filepath.Join(r.pluginsPath, string(plugin.Tier), id)
	report := &SecurityReport{
		ScannedAt: time.Now(),
		Score:     100,
	}

	filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)

		dangerousPatterns := []struct {
			pattern string
			issue   string
			score   int
		}{
			{"rm -rf /", "dangerous_rm", -50},
			{"curl | bash", "curl_pipe_bash", -30},
			{"> /dev/sd", "direct_disk_write", -50},
			{"wget | sh", "wget_pipe_sh", -30},
			{"eval(", "eval_usage", -20},
			{"__import__", "dynamic_import", -10},
			{"exec(", "exec_usage", -10},
		}

		for _, dp := range dangerousPatterns {
			if strings.Contains(content, dp.pattern) {
				report.DangerousCmds = true
				report.Issues = append(report.Issues, fmt.Sprintf("%s: found %s in %s", dp.issue, dp.pattern, filepath.Base(path)))
				report.Score += dp.score
			}
		}

		injectionPatterns := []string{
			"ignore previous instructions",
			"you are now",
			"new instructions:",
			"system prompt:",
		}
		for _, ip := range injectionPatterns {
			if strings.Contains(strings.ToLower(content), ip) {
				report.PromptInjection = true
				report.Issues = append(report.Issues, fmt.Sprintf("possible_prompt_injection: '%s' in %s", ip, filepath.Base(path)))
				report.Score -= 20
			}
		}

		return nil
	})

	if report.Score < 0 {
		report.Score = 0
	}
	if report.Score > 100 {
		report.Score = 100
	}

	plugin.SecurityScan = report
	plugin.TrustScore = float64(report.Score) / 100.0

	destDir := filepath.Join(r.pluginsPath, string(plugin.Tier), id)
	data, _ := json.MarshalIndent(plugin, "", "  ")
	os.WriteFile(filepath.Join(destDir, "plugin.json"), data, 0644)

	return report, nil
}

func (r *PluginRegistry) Search(query string) []*PluginMetadata {
	queryLower := strings.ToLower(query)
	var results []*PluginMetadata
	for _, p := range r.plugins {
		if strings.Contains(strings.ToLower(p.Name), queryLower) ||
			strings.Contains(strings.ToLower(p.Description), queryLower) {
			results = append(results, p)
		}
		for _, tag := range p.Tags {
			if strings.Contains(strings.ToLower(tag), queryLower) {
				results = append(results, p)
				break
			}
		}
	}
	return results
}

func (r *PluginRegistry) GetStats() map[string]interface{} {
	tierCounts := make(map[string]int)
	typeCounts := make(map[string]int)
	total := len(r.plugins)
	enabled := 0
	for _, p := range r.plugins {
		tierCounts[string(p.Tier)]++
		typeCounts[string(p.Type)]++
		if p.Enabled {
			enabled++
		}
	}
	return map[string]interface{}{
		"total":   total,
		"enabled": enabled,
		"tiers":   tierCounts,
		"types":   typeCounts,
	}
}
