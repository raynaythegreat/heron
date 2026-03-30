package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/raynaythegreat/heron/pkg/agent/coordination"
)

func (h *Handler) getCoordinator() *coordination.Coordinator {
	h.coordinatorOnce.Do(func() {
		h.coordinator = coordination.NewCoordinator(h.configPath)
	})
	return h.coordinator
}

func (h *Handler) registerCoordinationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/coordination/tasks", h.handleListTasks)
	mux.HandleFunc("POST /api/coordination/tasks", h.handleCreateTask)
	mux.HandleFunc("GET /api/coordination/tasks/{id}", h.handleGetTask)
	mux.HandleFunc("PUT /api/coordination/tasks/{id}", h.handleUpdateTask)
	mux.HandleFunc("POST /api/coordination/tasks/{id}/complete", h.handleCompleteTask)
	mux.HandleFunc("GET /api/coordination/agents", h.handleListAgents)
	mux.HandleFunc("POST /api/coordination/messages", h.handleSendMessage)
	mux.HandleFunc("GET /api/coordination/inbox/{agentId}", h.handleGetInbox)
	mux.HandleFunc("GET /api/coordination/stats", h.handleCoordinationStats)
}

func (h *Handler) handleListTasks(w http.ResponseWriter, r *http.Request) {
	coord := h.getCoordinator()
	tasks, err := coord.ListPendingTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func (h *Handler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		AgentID     string      `json:"agent_id"`
		Priority    int         `json:"priority"`
		BlockedBy   []string    `json:"blocked_by"`
		Input       interface{} `json:"input"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	coord := h.getCoordinator()
	task := &coordination.Task{
		Name:        body.Name,
		Description: body.Description,
		AgentID:     body.AgentID,
		Priority:    body.Priority,
		BlockedBy:   body.BlockedBy,
		Input:       body.Input,
	}
	if err := coord.CreateTask(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	coord := h.getCoordinator()
	task, err := coord.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string      `json:"status"`
		Output interface{} `json:"output"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	coord := h.getCoordinator()
	if err := coord.UpdateTaskStatus(id, coordination.TaskStatus(body.Status), body.Output, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task, _ := coord.GetTask(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Handler) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Output interface{} `json:"output"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	coord := h.getCoordinator()
	if err := coord.CompleteTask(id, body.Output); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "status": "completed"})
}

func (h *Handler) handleListAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"agents": []interface{}{}})
}

func (h *Handler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FromAgent string `json:"from_agent"`
		ToAgent   string `json:"to_agent"`
		TaskID    string `json:"task_id"`
		Type      string `json:"type"`
		Content   string `json:"content"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	coord := h.getCoordinator()
	msg := &coordination.AgentMessage{
		FromAgent: body.FromAgent,
		ToAgent:   body.ToAgent,
		TaskID:    body.TaskID,
		Type:      body.Type,
		Content:   body.Content,
	}
	if err := coord.SendMessage(msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func (h *Handler) handleGetInbox(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("agentId")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	coord := h.getCoordinator()
	messages, err := coord.ReceiveMessages(agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(messages) > limit {
		messages = messages[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_id": agentID,
		"limit":    limit,
		"messages": messages,
	})
}

func (h *Handler) handleCoordinationStats(w http.ResponseWriter, r *http.Request) {
	coord := h.getCoordinator()
	tasks, err := coord.ListPendingTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending_tasks": len(tasks),
		"in_progress":   0,
		"completed":     0,
		"active_agents": 0,
		"pending_waves": 0,
	})
}
