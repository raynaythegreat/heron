export interface HookProfile {
  profile: "minimal" | "standard" | "strict"
}

export interface HookEvent {
  event: string
  description: string
}

export async function getHookProfile(): Promise<HookProfile> {
  const res = await fetch("/api/hooks/profile")
  return res.json()
}

export async function setHookProfile(profile: string): Promise<HookProfile> {
  const res = await fetch("/api/hooks/profile", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ profile }),
  })
  return res.json()
}

export async function getHookEvents(): Promise<HookEvent[]> {
  const res = await fetch("/api/hooks/events")
  return res.json()
}

export async function getDisabledHooks(): Promise<string[]> {
  const res = await fetch("/api/hooks/disabled")
  return res.json()
}

export async function setDisabledHooks(hooks: string[]): Promise<{ disabled: string[] }> {
  const res = await fetch("/api/hooks/disabled", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ hooks }),
  })
  return res.json()
}
