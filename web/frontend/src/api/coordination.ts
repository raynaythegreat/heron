export interface Task {
  id: string
  name: string
  description: string
  agent_id: string
  status: "pending" | "in_progress" | "completed" | "failed" | "blocked"
  priority: number
  wave: number
  blocked_by: string[]
  assignee?: string
  input?: unknown
  output?: unknown
  worktree_path?: string
  created_at: string
  started_at?: string
  completed_at?: string
  error?: string
}

export interface AgentMessage {
  id: string
  from_agent: string
  to_agent: string
  task_id?: string
  type: string
  content: string
  timestamp: string
}

export async function listTasks(): Promise<{ tasks: Task[] }> {
  const res = await fetch("/api/coordination/tasks")
  return res.json()
}

export async function createTask(data: {
  name: string
  description?: string
  agent_id?: string
  priority?: number
  blocked_by?: string[]
  input?: unknown
}): Promise<{ status: string }> {
  const res = await fetch("/api/coordination/tasks", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function getTask(id: string): Promise<Task> {
  const res = await fetch(`/api/coordination/tasks/${id}`)
  return res.json()
}

export async function updateTask(id: string, data: { status: string; output?: unknown }): Promise<Task> {
  const res = await fetch(`/api/coordination/tasks/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function completeTask(id: string, output?: unknown): Promise<{ status: string }> {
  const res = await fetch(`/api/coordination/tasks/${id}/complete`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ output }),
  })
  return res.json()
}

export async function listAgents(): Promise<{ agents: unknown[] }> {
  const res = await fetch("/api/coordination/agents")
  return res.json()
}

export async function sendMessage(data: {
  from_agent: string
  to_agent: string
  task_id?: string
  type: string
  content: string
}): Promise<{ status: string }> {
  const res = await fetch("/api/coordination/messages", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  })
  return res.json()
}

export async function getInbox(agentId: string, limit?: number): Promise<{ messages: AgentMessage[] }> {
  const params = new URLSearchParams()
  if (limit) params.set("limit", String(limit))
  const res = await fetch(`/api/coordination/inbox/${agentId}?${params}`)
  return res.json()
}

export async function getCoordinationStats(): Promise<Record<string, number>> {
  const res = await fetch("/api/coordination/stats")
  return res.json()
}
