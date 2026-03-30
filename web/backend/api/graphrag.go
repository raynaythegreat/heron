package api

import (
	"encoding/json"
	"net/http"

	"github.com/raynaythegreat/heron/pkg/knowledge/graphrag"
)

func (h *Handler) getGraphEngine() *graphrag.GraphRAGEngine {
	h.graphEngineOnce.Do(func() {
		h.graphEngine = graphrag.NewGraphRAGEngine(h.configPath)
	})
	return h.graphEngine
}

func (h *Handler) registerGraphRAGRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/knowledge/graph", h.handleGetGraph)
	mux.HandleFunc("GET /api/knowledge/graph/search", h.handleGraphSearch)
	mux.HandleFunc("POST /api/knowledge/graph/scan", h.handleGraphScan)
	mux.HandleFunc("POST /api/knowledge/graph/diff-impact", h.handleDiffImpact)
	mux.HandleFunc("GET /api/knowledge/graph/stats", h.handleGraphStats)
}

func (h *Handler) handleGetGraph(w http.ResponseWriter, r *http.Request) {
	engine := h.getGraphEngine()
	data, err := engine.ExportForFrontend()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) handleGraphSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	engine := h.getGraphEngine()
	results := engine.Search(query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
	})
}

func (h *Handler) handleGraphScan(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Paths []string `json:"paths"`
		Force bool     `json:"force"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "scan_initiated"})
}

func (h *Handler) handleDiffImpact(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChangedFiles []string `json:"changed_files"`
	}
	if err := decodeJSON(r, &body); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	engine := h.getGraphEngine()
	impact := engine.DiffImpactAnalysis(body.ChangedFiles)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"changed_files": body.ChangedFiles,
		"impact":        impact,
	})
}

func (h *Handler) handleGraphStats(w http.ResponseWriter, r *http.Request) {
	engine := h.getGraphEngine()
	stats := engine.GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
