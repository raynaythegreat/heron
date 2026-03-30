package api

import (
	"encoding/json"
	"net/http"

	"github.com/raynaythegreat/heron/pkg/agent/learning"
)

func (h *Handler) getInstinctStore() *learning.InstinctStore {
	h.instinctStoreOnce.Do(func() {
		h.instinctStore = learning.NewInstinctStore()
	})
	return h.instinctStore
}

func (h *Handler) registerInstinctRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/instincts", h.handleListInstincts)
	mux.HandleFunc("GET /api/instincts/search", h.handleSearchInstincts)
	mux.HandleFunc("POST /api/instincts", h.handleCreateInstinct)
	mux.HandleFunc("POST /api/instincts/{id}/use", h.handleRecordInstinctUsage)
	mux.HandleFunc("POST /api/instincts/{id}/export", h.handleExportInstinct)
}

func (h *Handler) handleListInstincts(w http.ResponseWriter, r *http.Request) {
	store := h.getInstinctStore()
	if store == nil {
		http.Error(w, "instinct store not available", http.StatusInternalServerError)
		return
	}

	results, err := store.SearchInstincts("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"instincts": results})
}

func (h *Handler) handleSearchInstincts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	store := h.getInstinctStore()
	if store == nil {
		http.Error(w, "instinct store not available", http.StatusInternalServerError)
		return
	}

	results, err := store.SearchInstincts(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
	})
}

func (h *Handler) handleCreateInstinct(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Trigger     string   `json:"trigger"`
		Solution    string   `json:"solution"`
		Tags        []string `json:"tags"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	store := h.getInstinctStore()
	if store == nil {
		http.Error(w, "instinct store not available", http.StatusInternalServerError)
		return
	}

	instinct, err := store.ExtractInstinct(body.Name, body.Description, body.Trigger, body.Solution, body.Tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(instinct)
}

func (h *Handler) handleRecordInstinctUsage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store := h.getInstinctStore()
	if store == nil {
		http.Error(w, "instinct store not available", http.StatusInternalServerError)
		return
	}

	if err := store.RecordUsage(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": "recorded"})
}

func (h *Handler) handleExportInstinct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store := h.getInstinctStore()
	if store == nil {
		http.Error(w, "instinct store not available", http.StatusInternalServerError)
		return
	}

	skillMD, err := store.ExportToSkill(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       id,
		"skill_md": skillMD,
	})
}
