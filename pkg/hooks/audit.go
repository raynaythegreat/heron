package hooks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditHook struct {
	mu      sync.Mutex
	logPath string
	buffer  []map[string]interface{}
	maxBuf  int
}

func NewAuditHook() *AuditHook {
	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".heron", "logs")
	os.MkdirAll(logDir, 0755)
	return &AuditHook{
		logPath: filepath.Join(logDir, "audit.jsonl"),
		buffer:  make([]map[string]interface{}, 0, 100),
		maxBuf:  100,
	}
}

func (h *AuditHook) Name() string  { return "audit" }
func (h *AuditHook) Priority() int { return 10 }
func (h *AuditHook) Events() []HookEvent {
	return []HookEvent{
		EventPreToolUse, EventPostToolUse,
		EventPreCommand, EventPostCommand,
		EventPreFileWrite, EventPreFileRead,
		EventAgentSpawn, EventAgentComplete,
	}
}

func (h *AuditHook) Execute(ctx context.Context, hookCtx *HookContext) (*HookResult, error) {
	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339Nano),
		"event":     string(hookCtx.Event),
		"session":   hookCtx.SessionID,
		"agent":     hookCtx.AgentID,
		"tool":      hookCtx.ToolName,
	}

	if hookCtx.Input != nil {
		if m, ok := hookCtx.Input.(map[string]interface{}); ok {
			entry["input_summary"] = summarizeMap(m)
		} else {
			entry["input"] = hookCtx.Input
		}
	}

	h.mu.Lock()
	h.buffer = append(h.buffer, entry)
	if len(h.buffer) >= h.maxBuf {
		h.flush()
	}
	h.mu.Unlock()

	return &HookResult{Allowed: true}, nil
}

func (h *AuditHook) flush() {
	f, err := os.OpenFile(h.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	for _, entry := range h.buffer {
		encoder.Encode(entry)
	}
	h.buffer = h.buffer[:0]
}

func summarizeMap(m map[string]interface{}) map[string]interface{} {
	summary := make(map[string]interface{})
	for k, v := range m {
		if s, ok := v.(string); ok && len(s) > 200 {
			summary[k] = s[:200] + "..."
		} else {
			summary[k] = v
		}
	}
	return summary
}
