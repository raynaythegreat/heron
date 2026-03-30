package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/raynaythegreat/heron/pkg/memory/progressive"
)

func (h *Handler) getMemoryStore() *progressive.ProgressiveMemoryStore {
	h.memoryStoreOnce.Do(func() {
		store, err := progressive.NewProgressiveMemoryStore(h.configPath)
		if err == nil {
			h.memoryStore = store
		}
	})
	return h.memoryStore
}

func (h *Handler) registerMemoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/memory/search", h.handleMemorySearch)
	mux.HandleFunc("GET /api/memory/timeline", h.handleMemoryTimeline)
	mux.HandleFunc("POST /api/memory/store", h.handleMemoryStore)
	mux.HandleFunc("GET /api/memory/stats", h.handleMemoryStats)
	mux.HandleFunc("DELETE /api/memory/{id}", h.handleMemoryDelete)
}

func (h *Handler) handleMemorySearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	layerStr := r.URL.Query().Get("layer")
	limitStr := r.URL.Query().Get("limit")

	layer := progressive.LayerDetail
	switch layerStr {
	case "1", "search":
		layer = progressive.LayerSearch
	case "2", "timeline":
		layer = progressive.LayerTimeline
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if query == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	store := h.getMemoryStore()
	if store == nil {
		http.Error(w, "memory store not available", http.StatusInternalServerError)
		return
	}

	result, err := store.Search(r.Context(), query, layer, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) handleMemoryTimeline(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	store := h.getMemoryStore()
	if store == nil {
		http.Error(w, "memory store not available", http.StatusInternalServerError)
		return
	}

	entries, err := store.GetTimeline(sessionID, time.Time{}, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": sessionID,
		"limit":      limit,
		"entries":    entries,
	})
}

func (h *Handler) handleMemoryStore(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID string                 `json:"session_id"`
		Type      string                 `json:"type"`
		Summary   string                 `json:"summary"`
		Content   string                 `json:"content,omitempty"`
		Tags      []string               `json:"tags,omitempty"`
		Metadata  map[string]interface{} `json:"metadata,omitempty"`
		IsPrivate bool                   `json:"is_private"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	store := h.getMemoryStore()
	if store == nil {
		http.Error(w, "memory store not available", http.StatusInternalServerError)
		return
	}

	entry := &progressive.MemoryEntry{
		Layer:     progressive.LayerDetail,
		SessionID: body.SessionID,
		Timestamp: time.Now(),
		Type:      body.Type,
		Summary:   body.Summary,
		Content:   body.Content,
		Tags:      body.Tags,
		Metadata:  body.Metadata,
		IsPrivate: body.IsPrivate,
	}

	if err := store.Store(entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "stored", "id": entry.ID})
}

func (h *Handler) handleMemoryStats(w http.ResponseWriter, r *http.Request) {
	store := h.getMemoryStore()
	if store == nil {
		http.Error(w, "memory store not available", http.StatusInternalServerError)
		return
	}

	result, err := store.Search(context.Background(), "*", progressive.LayerDetail, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	byLayer := map[string]int{}
	for _, e := range result.Entries {
		switch e.Layer {
		case progressive.LayerSearch:
			byLayer["search"]++
		case progressive.LayerTimeline:
			byLayer["timeline"]++
		case progressive.LayerDetail:
			byLayer["detail"]++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_entries": result.TotalFound,
		"by_layer":      byLayer,
		"by_type":       map[string]int{},
	})
}

func (h *Handler) handleMemoryDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	store := h.getMemoryStore()
	if store == nil {
		http.Error(w, "memory store not available", http.StatusInternalServerError)
		return
	}

	if err := store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"deleted": id})
}
