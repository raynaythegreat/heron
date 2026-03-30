import {
  IconCircleDot,
  IconPlus,
  IconShieldCheck,
  IconTrash,
  IconRotate,
} from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { toast } from "sonner"

import {
  type HookEvent,
  getDisabledHooks,
  getHookEvents,
  getHookProfile,
  setDisabledHooks,
  setHookProfile,
} from "@/api/hooks"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { cn } from "@/lib/utils"

interface PermissionRule {
  id: string
  pattern: string
  action: "allow" | "ask" | "deny"
}

const PROFILES = [
  {
    key: "minimal" as const,
    title: "Minimal",
    description: "Only security hooks enabled",
    tagline: "Fewest prompts, fastest",
  },
  {
    key: "standard" as const,
    title: "Standard",
    description: "Security + permissions + audit",
    tagline: "Balanced",
  },
  {
    key: "strict" as const,
    title: "Strict",
    description: "All hooks enabled including memory",
    tagline: "Most secure, most prompts",
  },
] as const

const HOOK_EVENTS: Record<string, HookEvent[]> = {
  Session: [
    { event: "session.start", description: "Fired when a new session is created" },
    { event: "session.end", description: "Fired when a session ends" },
  ],
  Tool: [
    { event: "tool.before", description: "Before a tool is executed" },
    { event: "tool.after", description: "After a tool completes" },
    { event: "tool.error", description: "When a tool execution fails" },
  ],
  Command: [
    { event: "command.before", description: "Before a command is run" },
    { event: "command.after", description: "After a command completes" },
  ],
  File: [
    { event: "file.read", description: "When a file is read" },
    { event: "file.write", description: "When a file is written" },
    { event: "file.delete", description: "When a file is deleted" },
  ],
  Agent: [
    { event: "agent.think", description: "When the agent reasons" },
    { event: "agent.act", description: "When the agent takes an action" },
    { event: "agent.error", description: "When the agent encounters an error" },
  ],
  Skill: [
    { event: "skill.invoke", description: "When a skill is invoked" },
    { event: "skill.complete", description: "When a skill finishes" },
  ],
  Memory: [
    { event: "memory.store", description: "When data is stored in memory" },
    { event: "memory.recall", description: "When data is recalled from memory" },
  ],
}

const DEFAULT_RULES: PermissionRule[] = [
  { id: "1", pattern: "*.env", action: "deny" },
  { id: "2", pattern: "rm -rf", action: "deny" },
  { id: "3", pattern: "/tmp/*", action: "allow" },
]

export function HooksPage() {
  const queryClient = useQueryClient()
  const [rules, setRules] = useState<PermissionRule[]>(DEFAULT_RULES)
  const [newPattern, setNewPattern] = useState("")
  const [newAction, setNewAction] = useState<"allow" | "ask" | "deny">("ask")
  const [localDisabled, setLocalDisabled] = useState<Set<string>>(new Set())

  const { data: profile } = useQuery({
    queryKey: ["hookProfile"],
    queryFn: getHookProfile,
  })

  const { data: disabledHooks } = useQuery({
    queryKey: ["disabledHooks"],
    queryFn: getDisabledHooks,
  })

  const { data: hookEvents } = useQuery({
    queryKey: ["hookEvents"],
    queryFn: getHookEvents,
  })

  void hookEvents

  const profileMutation = useMutation({
    mutationFn: (p: string) => setHookProfile(p),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["hookProfile"] })
      toast.success("Profile updated")
    },
    onError: () => {
      toast.error("Failed to update profile")
    },
  })

  const disabledMutation = useMutation({
    mutationFn: (hooks: string[]) => setDisabledHooks(hooks),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["disabledHooks"] })
      toast.success("Disabled hooks updated")
    },
    onError: () => {
      toast.error("Failed to update disabled hooks")
    },
  })

  const allDisabled = new Set([
    ...(disabledHooks ?? []),
    ...localDisabled,
  ])

  const toggleHookEvent = (eventName: string, checked: boolean) => {
    setLocalDisabled((prev) => {
      const next = new Set(prev)
      if (checked) {
        next.delete(eventName)
      } else {
        next.add(eventName)
      }
      const fullList = [
        ...(disabledHooks ?? []).filter((h) => !localDisabled.has(h)),
        ...Array.from(next),
      ]
      disabledMutation.mutate(fullList)
      return next
    })
  }

  const reEnableHook = (hookName: string) => {
    setLocalDisabled((prev) => {
      const next = new Set(prev)
      next.delete(hookName)
      const fullList = [
        ...(disabledHooks ?? []).filter((h) => h !== hookName),
        ...Array.from(next),
      ]
      disabledMutation.mutate(fullList)
      return next
    })
  }

  const addRule = () => {
    if (!newPattern.trim()) return
    const rule: PermissionRule = {
      id: crypto.randomUUID(),
      pattern: newPattern.trim(),
      action: newAction,
    }
    setRules((prev) => [...prev, rule])
    setNewPattern("")
  }

  const removeRule = (id: string) => {
    setRules((prev) => prev.filter((r) => r.id !== id))
  }

  const actionColor = (action: string) => {
    switch (action) {
      case "allow":
        return "bg-emerald-500/15 text-emerald-600 border-emerald-500/25"
      case "deny":
        return "bg-red-500/15 text-red-600 border-red-500/25"
      case "ask":
        return "bg-amber-500/15 text-amber-600 border-amber-500/25"
      default:
        return ""
    }
  }

  const categoryColors: Record<string, string> = {
    Session: "text-cyan-400",
    Tool: "text-teal-400",
    Command: "text-emerald-400",
    File: "text-green-400",
    Agent: "text-sky-400",
    Skill: "text-indigo-400",
    Memory: "text-cyan-400",
  }

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <PageHeader title="Hook Configuration" />
      <div className="flex-1 overflow-y-auto px-6 pb-8">
        <p className="text-muted-foreground mb-6 text-sm">
          Manage security hooks, permissions, and event handlers
        </p>

        <div className="mb-8">
          <h3 className="text-foreground/90 mb-4 text-sm font-medium">
            Profile
          </h3>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
            {PROFILES.map((p) => {
              const isSelected = profile?.profile === p.key
              return (
                <Card
                  key={p.key}
                  className={cn(
                    "cursor-pointer transition-all hover:shadow-md",
                    isSelected
                      ? "ring-2 ring-cyan-500/60 bg-cyan-500/5"
                      : "hover:ring-1 hover:ring-foreground/10",
                  )}
                  onClick={() => profileMutation.mutate(p.key)}
                >
                  <CardHeader className="pb-0">
                    <div className="flex items-center gap-3">
                      <IconCircleDot
                        className={cn(
                          "size-4 shrink-0",
                          isSelected ? "fill-cyan-500 text-cyan-500" : "text-muted-foreground",
                        )}
                      />
                      <CardTitle className="text-sm">{p.title}</CardTitle>
                    </div>
                  </CardHeader>
                  <CardContent>
                    <CardDescription className="text-xs">
                      {p.description}
                    </CardDescription>
                    <Badge
                      variant="outline"
                      className={cn(
                        "mt-2 text-[10px]",
                        isSelected
                          ? "border-cyan-500/30 bg-cyan-500/10 text-cyan-500"
                          : "text-muted-foreground",
                      )}
                    >
                      {p.tagline}
                    </Badge>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </div>

        <Separator className="mb-8" />

        <div className="mb-8">
          <h3 className="text-foreground/90 mb-4 text-sm font-medium">
            Hook Events
          </h3>
          <Card>
            <CardContent className="p-0">
              <div className="divide-y divide-border/50">
                {Object.entries(HOOK_EVENTS).map(([category, events]) => (
                  <div key={category}>
                    <div className="bg-muted/30 px-4 py-2">
                      <span
                        className={cn(
                          "text-xs font-semibold uppercase tracking-wider",
                          categoryColors[category] ?? "text-muted-foreground",
                        )}
                      >
                        {category}
                      </span>
                    </div>
                    {events.map((evt) => {
                      const isDisabled = allDisabled.has(evt.event)
                      return (
                        <div
                          key={evt.event}
                          className="flex items-center justify-between px-4 py-2.5"
                        >
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-3">
                              <code className="bg-muted/60 rounded px-1.5 py-0.5 text-xs font-mono text-foreground/80">
                                {evt.event}
                              </code>
                              <span className="text-muted-foreground truncate text-xs">
                                {evt.description}
                              </span>
                            </div>
                          </div>
                          <Switch
                            checked={!isDisabled}
                            onCheckedChange={(checked) =>
                              toggleHookEvent(evt.event, checked)
                            }
                            size="sm"
                          />
                        </div>
                      )
                    })}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        <Separator className="mb-8" />

        <div className="mb-8">
          <h3 className="text-foreground/90 mb-4 text-sm font-medium">
            Permission Rules
          </h3>
          <Card>
            <CardContent className="space-y-3 pt-6">
              <div className="grid grid-cols-[1fr_auto_auto] items-center gap-3">
                <Input
                  placeholder="Pattern (e.g. *.env)"
                  value={newPattern}
                  onChange={(e) => setNewPattern(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && addRule()}
                  className="h-8 text-xs"
                />
                <Select value={newAction} onValueChange={(v) => setNewAction(v as "allow" | "ask" | "deny")}>
                  <SelectTrigger size="sm" className="h-8 w-24 text-xs">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="allow">Allow</SelectItem>
                    <SelectItem value="ask">Ask</SelectItem>
                    <SelectItem value="deny">Deny</SelectItem>
                  </SelectContent>
                </Select>
                <Button
                  variant="outline"
                  size="sm"
                  className="h-8"
                  onClick={addRule}
                >
                  <IconPlus className="size-3.5" />
                </Button>
              </div>

              <div className="space-y-2">
                {rules.map((rule) => (
                  <div
                    key={rule.id}
                    className="flex items-center justify-between rounded-md border border-border/50 bg-muted/20 px-3 py-2"
                  >
                    <div className="flex items-center gap-3">
                      <code className="text-xs font-mono text-foreground/80">
                        {rule.pattern}
                      </code>
                      <span className="text-muted-foreground text-xs">→</span>
                      <Badge
                        variant="outline"
                        className={cn("text-[10px]", actionColor(rule.action))}
                      >
                        {rule.action}
                      </Badge>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive"
                      onClick={() => removeRule(rule.id)}
                    >
                      <IconTrash className="size-3.5" />
                    </Button>
                  </div>
                ))}
                {rules.length === 0 && (
                  <p className="text-muted-foreground py-4 text-center text-xs">
                    No permission rules configured
                  </p>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        <Separator className="mb-8" />

        <div>
          <h3 className="text-foreground/90 mb-4 text-sm font-medium">
            Disabled Hooks
          </h3>
          <Card>
            <CardContent className="pt-6">
              {allDisabled.size > 0 ? (
                <div className="space-y-2">
                  {Array.from(allDisabled).map((hookName) => (
                    <div
                      key={hookName}
                      className="flex items-center justify-between rounded-md border border-border/50 bg-destructive/5 px-3 py-2"
                    >
                      <code className="text-xs font-mono text-foreground/70">
                        {hookName}
                      </code>
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-7 text-[10px] text-cyan-600 hover:bg-cyan-500/10 hover:text-cyan-600"
                        onClick={() => reEnableHook(hookName)}
                      >
                        <IconRotate className="mr-1 size-3" />
                        Re-enable
                      </Button>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="flex flex-col items-center py-6 text-center">
                  <IconShieldCheck className="text-muted-foreground/30 mb-2 size-8" />
                  <p className="text-muted-foreground text-xs">
                    All hooks are enabled
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
