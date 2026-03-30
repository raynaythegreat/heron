export interface MemoryEntry {
  id: string
  layer: number
  session_id: string
  timestamp: string
  type: string
  summary: string
  content?: string
  tags?: string[]
  is_private?: boolean
}

export interface MemorySearchResult {
  entries: MemoryEntry[]
  total_found: number
  tokens_used: number
  layer_retrieved: number
}

export interface MemoryStats {
  total_entries: number
  by_layer: Record<string, number>
  by_type: Record<string, number>
}

export async function searchMemory(query: string, layer?: number, limit?: number): Promise<MemorySearchResult> {
  const params = new URLSearchParams({ q: query })
  if (layer) params.set("layer", String(layer))
  if (limit) params.set("limit", String(limit))
  const res = await fetch(`/api/memory/search?${params}`)
  return res.json()
}

export async function getMemoryTimeline(sessionId: string, limit?: number): Promise<{ entries: MemoryEntry[] }> {
  const params = new URLSearchParams({ session_id: sessionId })
  if (limit) params.set("limit", String(limit))
  const res = await fetch(`/api/memory/timeline?${params}`)
  return res.json()
}

export async function storeMemory(data: {
  session_id: string
  type: string
  summary: string
  content?: string
  tags?: string[]
  metadata?: Record<string, unknown>
  is_private?: boolean
}): Promise<{ status: string }> {
  const res = await fetch("/api/memory/store", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function getMemoryStats(): Promise<MemoryStats> {
  const res = await fetch("/api/memory/stats")
  return res.json()
}

export async function deleteMemory(id: string): Promise<{ deleted: string }> {
  const res = await fetch(`/api/memory/${id}`, { method: "DELETE" })
  return res.json()
}
