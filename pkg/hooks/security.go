package hooks

import (
	"context"
	"strings"
)

type SecurityHook struct{}

func NewSecurityHook() *SecurityHook { return &SecurityHook{} }

func (h *SecurityHook) Name() string  { return "security" }
func (h *SecurityHook) Priority() int { return 100 }
func (h *SecurityHook) Events() []HookEvent {
	return []HookEvent{EventPreCommand, EventPreFileWrite, EventPreFileRead, EventUserPrompt}
}

var dangerousPatterns = []string{
	"rm -rf /", "rm -rf /*", "mkfs", "dd if=", ":(){ :|:& };:",
	"chmod -R 777 /", "> /dev/sd", "shutdown", "reboot",
	" DROP ", " DELETE FROM ", " TRUNCATE ",
}

func (h *SecurityHook) Execute(ctx context.Context, hookCtx *HookContext) (*HookResult, error) {
	var input string
	switch v := hookCtx.Input.(type) {
	case string:
		input = v
	case map[string]interface{}:
		if cmd, ok := v["command"].(string); ok {
			input = cmd
		}
		if content, ok := v["content"].(string); ok {
			input = content
		}
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(input, pattern) {
			return &HookResult{
				Allowed: false,
				Message: "Security hook blocked dangerous command: " + pattern,
			}, nil
		}
	}

	return &HookResult{Allowed: true}, nil
}
