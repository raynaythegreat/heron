package hooks

import (
	"context"
	"time"
)

type MemoryHook struct{}

func NewMemoryHook() *MemoryHook { return &MemoryHook{} }

func (h *MemoryHook) Name() string  { return "memory" }
func (h *MemoryHook) Priority() int { return 50 }
func (h *MemoryHook) Events() []HookEvent {
	return []HookEvent{EventSessionStart, EventSessionEnd, EventUserPrompt, EventMemoryStore, EventMemoryRetrieve}
}

func (h *MemoryHook) Execute(ctx context.Context, hookCtx *HookContext) (*HookResult, error) {
	switch hookCtx.Event {
	case EventSessionStart:
		return &HookResult{
			Allowed: true,
			Metadata: map[string]string{
				"memory_action": "session_start",
				"session_id":    hookCtx.SessionID,
				"timestamp":     time.Now().Format(time.RFC3339),
			},
		}, nil
	case EventMemoryRetrieve:
		return &HookResult{
			Allowed:  true,
			Metadata: map[string]string{"memory_action": "retrieve"},
		}, nil
	}
	return &HookResult{Allowed: true}, nil
}
