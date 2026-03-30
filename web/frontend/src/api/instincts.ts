export interface Instinct {
  id: string
  name: string
  description: string
  trigger: string
  solution: string
  confidence: number
  times_used: number
  created_at: string
  last_used_at?: string
  tags?: string[]
  verified: boolean
}

export async function listInstincts(): Promise<{ instincts: Instinct[] }> {
  const res = await fetch("/api/instincts")
  return res.json()
}

export async function searchInstincts(query: string): Promise<{ query: string; results: Instinct[] }> {
  const res = await fetch(`/api/instincts/search?q=${encodeURIComponent(query)}`)
  return res.json()
}

export async function createInstinct(data: {
  name: string
  description: string
  trigger: string
  solution: string
  tags?: string[]
}): Promise<{ status: string }> {
  const res = await fetch("/api/instincts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function recordInstinctUsage(id: string): Promise<{ id: string; status: string }> {
  const res = await fetch(`/api/instincts/${id}/use`, { method: "POST" })
  return res.json()
}

export async function exportInstinct(id: string): Promise<{ id: string; skill_md: string }> {
  const res = await fetch(`/api/instincts/${id}/export`, { method: "POST" })
  return res.json()
}
