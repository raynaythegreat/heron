export interface Plugin {
  id: string
  name: string
  description: string
  version?: string
  tier: "system" | "curated" | "experimental" | "local"
  type?: "skill" | "plugin"
  author?: string
  tags?: string[]
  enabled: boolean
  trust_score: number
  scan_status: "clean" | "warning" | "unscanned" | "scanning"
  installed?: boolean
  security_scan?: {
    scanned_at: string
    prompt_injection: boolean
    malicious_code: boolean
    dangerous_cmds: boolean
    score: number
    issues?: string[]
  }
}

export async function listPlugins(tier?: string): Promise<{ tier: string; plugins: Plugin[] }> {
  const params = tier ? `?tier=${tier}` : ""
  const res = await fetch(`/api/plugins${params}`)
  return res.json()
}

export async function getPlugin(id: string): Promise<Plugin> {
  const res = await fetch(`/api/plugins/${id}`)
  return res.json()
}

export async function installPlugin(source: string, tier?: string): Promise<{ status: string }> {
  const res = await fetch("/api/plugins/install", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ source, tier }),
  })
  return res.json()
}

export async function uninstallPlugin(id: string): Promise<{ deleted: string }> {
  const res = await fetch(`/api/plugins/${id}`, { method: "DELETE" })
  return res.json()
}

export async function togglePlugin(id: string, enabled: boolean): Promise<{ id: string; enabled: boolean }> {
  const res = await fetch(`/api/plugins/${id}/toggle`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ enabled }),
  })
  return res.json()
}

export async function scanPlugin(id: string): Promise<{ id: string; security: Plugin["security_scan"] }> {
  const res = await fetch(`/api/plugins/${id}/scan`, { method: "POST" })
  return res.json()
}

export async function getPluginStats(): Promise<Record<string, number>> {
  const res = await fetch("/api/plugins/stats")
  return res.json()
}

export async function searchPlugins(query: string): Promise<{ query: string; results: Plugin[] }> {
  const res = await fetch(`/api/plugins/search?q=${encodeURIComponent(query)}`)
  return res.json()
}
