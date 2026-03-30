export interface GraphNode {
  id: string
  type: string
  name: string
  path?: string
  language?: string
}

export interface GraphEdge {
  source: string
  target: string
  type: string
}

export interface CodeGraph {
  nodes: GraphNode[]
  edges: GraphEdge[]
  stats: Record<string, unknown>
}

export async function getGraph(): Promise<CodeGraph> {
  const res = await fetch("/api/knowledge/graph")
  return res.json()
}

export async function searchGraph(query: string): Promise<{ query: string; results: GraphNode[] }> {
  const res = await fetch(`/api/knowledge/graph/search?q=${encodeURIComponent(query)}`)
  return res.json()
}

export async function scanGraph(paths: string[], force?: boolean): Promise<{ status: string }> {
  const res = await fetch("/api/knowledge/graph/scan", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ paths, force }),
  })
  return res.json()
}

export async function getDiffImpact(changedFiles: string[]): Promise<{ changed_files: string[]; impact: Record<string, number> }> {
  const res = await fetch("/api/knowledge/graph/diff-impact", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ changed_files: changedFiles }),
  })
  return res.json()
}

export async function getGraphStats(): Promise<Record<string, unknown>> {
  const res = await fetch("/api/knowledge/graph/stats")
  return res.json()
}
