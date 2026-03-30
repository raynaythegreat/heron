package api

import (
	"encoding/json"
	"net/http"

	"github.com/raynaythegreat/heron/pkg/hooks"
)

func (h *Handler) getHookManager() *hooks.HookManager {
	h.hookManagerOnce.Do(func() {
		h.hookManager = hooks.NewHookManager(h.configPath)
	})
	return h.hookManager
}

func (h *Handler) registerHooksRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/hooks/profile", h.handleGetHookProfile)
	mux.HandleFunc("PUT /api/hooks/profile", h.handleSetHookProfile)
	mux.HandleFunc("GET /api/hooks/events", h.handleGetHookEvents)
	mux.HandleFunc("GET /api/hooks/disabled", h.handleGetDisabledHooks)
	mux.HandleFunc("PUT /api/hooks/disabled", h.handleSetDisabledHooks)
}

func (h *Handler) handleGetHookProfile(w http.ResponseWriter, r *http.Request) {
	mgr := h.getHookManager()
	cfg := mgr.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"profile": cfg.Profile})
}

func (h *Handler) handleSetHookProfile(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Profile string `json:"profile"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if body.Profile != "minimal" && body.Profile != "standard" && body.Profile != "strict" {
		http.Error(w, "invalid profile", http.StatusBadRequest)
		return
	}
	mgr := h.getHookManager()
	mgr.SetProfile(body.Profile)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"profile": body.Profile})
}

func (h *Handler) handleGetHookEvents(w http.ResponseWriter, r *http.Request) {
	events := []map[string]interface{}{
		{"event": "session_start", "description": "Fired when a new session begins"},
		{"event": "session_end", "description": "Fired when a session ends"},
		{"event": "user_prompt", "description": "Fired when user sends a message"},
		{"event": "pre_tool_use", "description": "Fired before a tool is executed"},
		{"event": "post_tool_use", "description": "Fired after a tool completes"},
		{"event": "pre_command", "description": "Fired before shell command execution"},
		{"event": "post_command", "description": "Fired after shell command completes"},
		{"event": "pre_file_write", "description": "Fired before file write operations"},
		{"event": "pre_file_read", "description": "Fired before file read operations"},
		{"event": "agent_spawn", "description": "Fired when a sub-agent is spawned"},
		{"event": "agent_complete", "description": "Fired when a sub-agent completes"},
		{"event": "skill_activate", "description": "Fired when a skill is activated"},
		{"event": "memory_store", "description": "Fired when a memory is stored"},
		{"event": "memory_retrieve", "description": "Fired when memory is retrieved"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (h *Handler) handleGetDisabledHooks(w http.ResponseWriter, r *http.Request) {
	mgr := h.getHookManager()
	cfg := mgr.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg.DisabledHooks)
}

func (h *Handler) handleSetDisabledHooks(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Hooks []string `json:"hooks"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	mgr := h.getHookManager()
	mgr.SetDisabledHooks(body.Hooks)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"disabled": body.Hooks})
}
