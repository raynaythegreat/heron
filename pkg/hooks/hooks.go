package hooks

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type HookEvent string

const (
	EventSessionStart    HookEvent = "session_start"
	EventSessionEnd      HookEvent = "session_end"
	EventUserPrompt      HookEvent = "user_prompt"
	EventPreToolUse      HookEvent = "pre_tool_use"
	EventPostToolUse     HookEvent = "post_tool_use"
	EventPreCommand      HookEvent = "pre_command"
	EventPostCommand     HookEvent = "post_command"
	EventPreFileWrite    HookEvent = "pre_file_write"
	EventPreFileRead     HookEvent = "pre_file_read"
	EventAgentSpawn      HookEvent = "agent_spawn"
	EventAgentComplete   HookEvent = "agent_complete"
	EventSkillActivate   HookEvent = "skill_activate"
	EventSkillDeactivate HookEvent = "skill_deactivate"
	EventMemoryStore     HookEvent = "memory_store"
	EventMemoryRetrieve  HookEvent = "memory_retrieve"
)

type HookResult struct {
	Allowed  bool              `json:"allowed"`
	Modified interface{}       `json:"modified,omitempty"`
	Message  string            `json:"message,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type HookContext struct {
	Event      HookEvent       `json:"event"`
	ToolName   string          `json:"tool_name,omitempty"`
	Input      interface{}     `json:"input,omitempty"`
	Output     interface{}     `json:"output,omitempty"`
	SessionID  string          `json:"session_id,omitempty"`
	AgentID    string          `json:"agent_id,omitempty"`
	Timestamp  time.Time       `json:"timestamp"`
	Permission PermissionLevel `json:"permission"`
}

type PermissionLevel int

const (
	PermissionAllow PermissionLevel = iota
	PermissionAsk
	PermissionDeny
)

type Hook interface {
	Name() string
	Events() []HookEvent
	Priority() int
	Execute(ctx context.Context, hookCtx *HookContext) (*HookResult, error)
}

type HookManager struct {
	mu     sync.RWMutex
	hooks  map[HookEvent][]Hook
	config *HookConfig
}

type HookConfig struct {
	Profile         string            `json:"profile"`
	DisabledHooks   []string          `json:"disabled_hooks"`
	PermissionRules map[string]string `json:"permission_rules"`
	MaxRetries      int               `json:"max_retries"`
	Timeout         time.Duration     `json:"timeout"`
}

func NewHookManager(workspacePath string) *HookManager {
	hm := &HookManager{
		hooks: make(map[HookEvent][]Hook),
		config: &HookConfig{
			Profile:         "standard",
			DisabledHooks:   []string{},
			PermissionRules: make(map[string]string),
			MaxRetries:      3,
			Timeout:         30 * time.Second,
		},
	}
	hm.loadBuiltinHooks()
	return hm
}

func (hm *HookManager) loadBuiltinHooks() {
	hm.Register(NewSecurityHook())
	hm.Register(NewPermissionHook(hm.config))
	hm.Register(NewMemoryHook())
	hm.Register(NewAuditHook())
}

func (hm *HookManager) Register(hook Hook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	for _, event := range hook.Events() {
		hm.hooks[event] = append(hm.hooks[event], hook)
	}
	hm.sortHooks()
}

func (hm *HookManager) sortHooks() {
	for event := range hm.hooks {
		hooks := hm.hooks[event]
		for i := 0; i < len(hooks)-1; i++ {
			for j := i + 1; j < len(hooks); j++ {
				if hooks[j].Priority() > hooks[i].Priority() {
					hooks[i], hooks[j] = hooks[j], hooks[i]
				}
			}
		}
		hm.hooks[event] = hooks
	}
}

func (hm *HookManager) Fire(ctx context.Context, event HookEvent, hookCtx *HookContext) (*HookResult, error) {
	hm.mu.RLock()
	hooks := hm.hooks[event]
	hm.mu.RUnlock()

	if len(hooks) == 0 {
		return &HookResult{Allowed: true}, nil
	}

	hookCtx.Event = event
	hookCtx.Timestamp = time.Now()

	for _, hook := range hooks {
		skip := false
		for _, disabled := range hm.config.DisabledHooks {
			if hook.Name() == disabled {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		ctx, cancel := context.WithTimeout(ctx, hm.config.Timeout)
		result, err := hook.Execute(ctx, hookCtx)
		cancel()

		if err != nil {
			log.Error().Err(err).Str("hook", hook.Name()).Str("event", string(event)).Msg("hook execution failed")
			continue
		}

		if !result.Allowed {
			log.Info().Str("hook", hook.Name()).Str("event", string(event)).Msg("hook blocked action")
			return result, nil
		}

		if result.Modified != nil {
			hookCtx.Input = result.Modified
		}
	}

	return &HookResult{Allowed: true}, nil
}

func (hm *HookManager) GetConfig() *HookConfig {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.config
}

func (hm *HookManager) SetDisabledHooks(hooks []string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.config.DisabledHooks = hooks
}

func (hm *HookManager) SetProfile(profile string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	switch profile {
	case "minimal":
		hm.config.DisabledHooks = []string{"audit", "memory_store"}
	case "strict":
		hm.config.DisabledHooks = []string{}
	case "standard":
		hm.config.DisabledHooks = []string{}
	default:
		hm.config.DisabledHooks = []string{}
	}
	hm.config.Profile = profile
}
