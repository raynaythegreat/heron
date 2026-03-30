package hooks

import (
	"context"
	"strings"
)

type PermissionHook struct {
	config *HookConfig
}

func NewPermissionHook(config *HookConfig) *PermissionHook {
	return &PermissionHook{config: config}
}

func (h *PermissionHook) Name() string  { return "permission" }
func (h *PermissionHook) Priority() int { return 90 }
func (h *PermissionHook) Events() []HookEvent {
	return []HookEvent{EventPreCommand, EventPreFileWrite, EventPreFileRead, EventPreToolUse}
}

func (h *PermissionHook) Execute(ctx context.Context, hookCtx *HookContext) (*HookResult, error) {
	if h.config == nil {
		return &HookResult{Allowed: true}, nil
	}

	var toolName string
	switch v := hookCtx.Input.(type) {
	case string:
		toolName = v
	case map[string]interface{}:
		if cmd, ok := v["command"].(string); ok {
			toolName = cmd
		}
		if name, ok := v["tool"].(string); ok {
			toolName = name
		}
	}

	for pattern, level := range h.config.PermissionRules {
		if strings.Contains(toolName, pattern) {
			switch level {
			case "deny":
				return &HookResult{Allowed: false, Message: "Permission denied by rule: " + pattern}, nil
			case "ask":
				return &HookResult{Allowed: true, Message: "Permission rule matched (ask): " + pattern}, nil
			}
		}
	}

	return &HookResult{Allowed: true}, nil
}
