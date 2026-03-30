package graphrag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NodeType string

const (
	NodeTypeFile      NodeType = "file"
	NodeTypeFunction  NodeType = "function"
	NodeTypeClass     NodeType = "class"
	NodeTypeInterface NodeType = "interface"
	NodeTypeStruct    NodeType = "struct"
	NodeTypeVariable  NodeType = "variable"
	NodeTypeImport    NodeType = "import"
	NodeTypeModule    NodeType = "module"
)

type GraphNode struct {
	ID        string            `json:"id"`
	Type      NodeType          `json:"type"`
	Name      string            `json:"name"`
	Path      string            `json:"path,omitempty"`
	Language  string            `json:"language,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Embedding []float64         `json:"embedding,omitempty"`
}

type GraphEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Type   string `json:"type"`
	Label  string `json:"label,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

type CodeGraph struct {
	Nodes map[string]*GraphNode `json:"nodes"`
	Edges []GraphEdge           `json:"edges"`
}

type GraphRAGEngine struct {
	graph       *CodeGraph
	storagePath string
}

func NewGraphRAGEngine(workspacePath string) *GraphRAGEngine {
	home, _ := os.UserHomeDir()
	storagePath := filepath.Join(home, ".heron", "knowledge", "graph")
	os.MkdirAll(storagePath, 0755)

	engine := &GraphRAGEngine{
		graph: &CodeGraph{
			Nodes: make(map[string]*GraphNode),
			Edges: []GraphEdge{},
		},
		storagePath: storagePath,
	}

	engine.load()
	return engine
}

func (e *GraphRAGEngine) AddNode(node *GraphNode) {
	e.graph.Nodes[node.ID] = node
}

func (e *GraphRAGEngine) AddEdge(from, to, edgeType, label string, weight int) {
	e.graph.Edges = append(e.graph.Edges, GraphEdge{
		From:   from,
		To:     to,
		Type:   edgeType,
		Label:  label,
		Weight: weight,
	})
}

func (e *GraphRAGEngine) GetNode(id string) *GraphNode {
	return e.graph.Nodes[id]
}

func (e *GraphRAGEngine) GetConnectedNodes(nodeID string, edgeType string) []*GraphNode {
	var connected []*GraphNode
	for _, edge := range e.graph.Edges {
		if edgeType != "" && edge.Type != edgeType {
			continue
		}
		var targetID string
		if edge.From == nodeID {
			targetID = edge.To
		} else if edge.To == nodeID {
			targetID = edge.From
		} else {
			continue
		}
		if node, ok := e.graph.Nodes[targetID]; ok {
			connected = append(connected, node)
		}
	}
	return connected
}

func (e *GraphRAGEngine) GetImpact(nodeID string) map[string]int {
	impact := make(map[string]int)
	e.propagateImpact(nodeID, impact, 3, 1.0)
	return impact
}

func (e *GraphRAGEngine) propagateImpact(nodeID string, impact map[string]int, maxDepth int, weight float64) {
	if maxDepth <= 0 {
		return
	}
	for _, edge := range e.graph.Edges {
		var targetID string
		if edge.From == nodeID {
			targetID = edge.To
		} else if edge.To == nodeID {
			targetID = edge.From
		} else {
			continue
		}
		edgeWeight := edge.Weight
		if edgeWeight == 0 {
			edgeWeight = 1
		}
		impact[targetID] += int(float64(edgeWeight) * weight)
		e.propagateImpact(targetID, impact, maxDepth-1, weight*0.5)
	}
}

func (e *GraphRAGEngine) Search(query string) []*GraphNode {
	queryLower := strings.ToLower(query)
	var results []*GraphNode
	for _, node := range e.graph.Nodes {
		if strings.Contains(strings.ToLower(node.Name), queryLower) ||
			strings.Contains(strings.ToLower(node.Path), queryLower) {
			results = append(results, node)
		}
	}
	return results
}

func (e *GraphRAGEngine) DiffImpactAnalysis(changedFiles []string) map[string]float64 {
	impact := make(map[string]float64)
	for _, file := range changedFiles {
		nodeID := "file:" + file
		nodeImpact := e.GetImpact(nodeID)
		for affected, weight := range nodeImpact {
			if affected != nodeID {
				impact[affected] += float64(weight)
			}
		}
	}
	return impact
}

func (e *GraphRAGEngine) GetStats() map[string]interface{} {
	nodeTypes := make(map[NodeType]int)
	for _, node := range e.graph.Nodes {
		nodeTypes[node.Type]++
	}
	return map[string]interface{}{
		"total_nodes":  len(e.graph.Nodes),
		"total_edges":  len(e.graph.Edges),
		"node_types":   nodeTypes,
		"last_scanned": time.Now().Format(time.RFC3339),
		"storage_path": e.storagePath,
	}
}

func (e *GraphRAGEngine) Save() error {
	data, err := json.MarshalIndent(e.graph, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := filepath.Join(e.storagePath, "codegraph.tmp")
	finalPath := filepath.Join(e.storagePath, "codegraph.json")
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, finalPath)
}

func (e *GraphRAGEngine) load() {
	data, err := os.ReadFile(filepath.Join(e.storagePath, "codegraph.json"))
	if err != nil {
		return
	}
	json.Unmarshal(data, e.graph)
}

func (e *GraphRAGEngine) ExportForFrontend() (interface{}, error) {
	type FrontendNode struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Name     string `json:"name"`
		Path     string `json:"path,omitempty"`
		Language string `json:"language,omitempty"`
	}

	type FrontendEdge struct {
		From string `json:"source"`
		To   string `json:"target"`
		Type string `json:"type"`
	}

	nodes := make([]FrontendNode, 0, len(e.graph.Nodes))
	for _, n := range e.graph.Nodes {
		nodes = append(nodes, FrontendNode{
			ID:       n.ID,
			Type:     string(n.Type),
			Name:     n.Name,
			Path:     n.Path,
			Language: n.Language,
		})
	}

	edges := make([]FrontendEdge, 0, len(e.graph.Edges))
	for _, edge := range e.graph.Edges {
		edges = append(edges, FrontendEdge{
			From: edge.From,
			To:   edge.To,
			Type: edge.Type,
		})
	}

	return map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
		"stats": e.GetStats(),
	}, nil
}
