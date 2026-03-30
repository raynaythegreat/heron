import { IconFilter, IconLoader, IconPlus, IconRobot } from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"

import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Progress } from "@/components/ui/progress"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import {
  type AgentMessage,
  type Task,
  createTask,
  getCoordinationStats,
  getInbox,
  listAgents,
  listTasks,
} from "@/api/coordination"

const statusColors: Record<Task["status"], string> = {
  pending: "bg-yellow-500/15 text-yellow-600 border-yellow-500/25",
  in_progress: "bg-cyan-500/15 text-cyan-600 border-cyan-500/25",
  completed: "bg-green-500/15 text-green-600 border-green-500/25",
  failed: "bg-red-500/15 text-red-600 border-red-500/25",
  blocked: "bg-zinc-500/15 text-zinc-500 border-zinc-500/25",
}

const statusLabels: Record<Task["status"], string> = {
  pending: "Pending",
  in_progress: "In Progress",
  completed: "Completed",
  failed: "Failed",
  blocked: "Blocked",
}

function formatTimestamp(ts: string) {
  return new Date(ts).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <Card size="sm">
      <CardContent className="pb-2 pt-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className={`text-2xl font-semibold ${color}`}>{value}</p>
      </CardContent>
    </Card>
  )
}

function CreateTaskSheet({
  open,
  onClose,
}: {
  open: boolean
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [agentId, setAgentId] = useState("")
  const [priority, setPriority] = useState("3")

  const { data: agentsData } = useQuery({
    queryKey: ["coordination-agents"],
    queryFn: listAgents,
  })

  const agents = (agentsData?.agents ?? []) as Array<{ id: string; name?: string }>

  const mutation = useMutation({
    mutationFn: () =>
      createTask({
        name,
        description: description || undefined,
        agent_id: agentId || undefined,
        priority: Number(priority),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["coordination-tasks"] })
      queryClient.invalidateQueries({ queryKey: ["coordination-stats"] })
      setName("")
      setDescription("")
      setAgentId("")
      setPriority("3")
      onClose()
    },
  })

  return (
    <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>Create Task</SheetTitle>
          <SheetDescription>
            Add a new task to the coordination queue.
          </SheetDescription>
        </SheetHeader>
        <div className="flex flex-col gap-4 py-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="task-name">Name</Label>
            <Input
              id="task-name"
              placeholder="Task name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="task-desc">Description</Label>
            <Textarea
              id="task-desc"
              placeholder="Optional description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="task-agent">Agent</Label>
            <Select value={agentId} onValueChange={setAgentId}>
              <SelectTrigger id="task-agent">
                <SelectValue placeholder="Assign to agent (optional)" />
              </SelectTrigger>
              <SelectContent>
                {agents.map((a) => (
                  <SelectItem key={a.id} value={a.id}>
                    {a.name ?? a.id}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="task-priority">Priority</Label>
            <Select value={priority} onValueChange={setPriority}>
              <SelectTrigger id="task-priority">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">1 - Highest</SelectItem>
                <SelectItem value="2">2 - High</SelectItem>
                <SelectItem value="3">3 - Normal</SelectItem>
                <SelectItem value="4">4 - Low</SelectItem>
                <SelectItem value="5">5 - Lowest</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <SheetFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={() => mutation.mutate()}
            disabled={!name.trim() || mutation.isPending}
          >
            {mutation.isPending ? "Creating..." : "Create Task"}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}

function TasksTab() {
  const [createOpen, setCreateOpen] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: ["coordination-tasks"],
    queryFn: listTasks,
  })

  const tasks = data?.tasks ?? []

  return (
    <>
      <div className="mb-4 flex items-center justify-end">
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <IconPlus className="size-4" />
          Create Task
        </Button>
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          <IconLoader className="size-5 animate-spin mr-2" />
          <span className="text-sm">Loading tasks...</span>
        </div>
      )}

      {error && (
        <div className="py-12 text-center text-sm text-destructive">
          Failed to load tasks: {error.message}
        </div>
      )}

      {!isLoading && !error && tasks.length === 0 && (
        <div className="py-12 text-center text-sm text-muted-foreground">
          No tasks queued.{" "}
          <button
            onClick={() => setCreateOpen(true)}
            className="underline underline-offset-4 hover:text-foreground transition-colors"
          >
            Create a task
          </button>{" "}
          to get started.
        </div>
      )}

      {!isLoading && !error && tasks.length > 0 && (
        <div className="rounded-lg border">
          <div className="grid grid-cols-[1fr_120px_120px_80px_60px_140px] gap-2 border-b bg-muted/50 px-4 py-2 text-xs font-medium text-muted-foreground">
            <span>Name</span>
            <span>Status</span>
            <span>Agent</span>
            <span>Priority</span>
            <span>Wave</span>
            <span>Created</span>
          </div>
          {tasks.map((task) => (
            <div
              key={task.id}
              className="grid grid-cols-[1fr_120px_120px_80px_60px_140px] gap-2 border-b px-4 py-3 text-sm last:border-0 hover:bg-muted/30 transition-colors"
            >
              <div className="flex flex-col gap-0.5 min-w-0">
                <span className="font-medium truncate">{task.name}</span>
                {task.description && (
                  <span className="text-xs text-muted-foreground truncate">
                    {task.description}
                  </span>
                )}
              </div>
              <div>
                <Badge
                  variant="outline"
                  className={statusColors[task.status]}
                >
                  {statusLabels[task.status]}
                </Badge>
              </div>
              <span className="text-muted-foreground truncate">
                {task.assignee ?? task.agent_id ?? "—"}
              </span>
              <span>{task.priority}</span>
              <span>{task.wave}</span>
              <span className="text-xs text-muted-foreground">
                {formatTimestamp(task.created_at)}
              </span>
            </div>
          ))}
        </div>
      )}

      <CreateTaskSheet open={createOpen} onClose={() => setCreateOpen(false)} />
    </>
  )
}

function MessagesTab() {
  const [agentFilter, setAgentFilter] = useState("")

  const { data, isLoading, error } = useQuery({
    queryKey: ["coordination-agents"],
    queryFn: listAgents,
  })

  const agents = (data?.agents ?? []) as Array<{ id: string; name?: string }>

  const { data: inboxData } = useQuery({
    queryKey: ["coordination-inbox", agentFilter],
    queryFn: () => getInbox(agentFilter, 50),
    enabled: !!agentFilter,
  })

  const messages: AgentMessage[] = inboxData?.messages ?? []

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <IconFilter className="size-4 text-muted-foreground" />
        <Select value={agentFilter} onValueChange={setAgentFilter}>
          <SelectTrigger className="w-64">
            <SelectValue placeholder="Filter by agent..." />
          </SelectTrigger>
          <SelectContent>
            {agents.map((a) => (
              <SelectItem key={a.id} value={a.id}>
                {a.name ?? a.id}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          <IconLoader className="size-5 animate-spin mr-2" />
          <span className="text-sm">Loading agents...</span>
        </div>
      )}

      {error && (
        <div className="py-12 text-center text-sm text-destructive">
          Failed to load agents: {error.message}
        </div>
      )}

      {!agentFilter && (
        <div className="py-12 text-center text-sm text-muted-foreground">
          Select an agent to view their messages.
        </div>
      )}

      {agentFilter && messages.length === 0 && (
        <div className="py-12 text-center text-sm text-muted-foreground">
          No messages found for this agent.
        </div>
      )}

      {messages.length > 0 && (
        <ScrollArea className="h-[480px] rounded-lg border">
          <div className="flex flex-col">
            {messages.map((msg) => (
              <div
                key={msg.id}
                className="flex flex-col gap-1 border-b px-4 py-3 last:border-0 hover:bg-muted/30 transition-colors"
              >
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">
                    {msg.from_agent}
                  </span>
                  <span className="text-cyan-500">&rarr;</span>
                  <span className="font-medium text-foreground">
                    {msg.to_agent}
                  </span>
                  <span className="ml-auto">{formatTimestamp(msg.timestamp)}</span>
                </div>
                <p className="text-sm line-clamp-2">{msg.content}</p>
                {msg.type && (
                  <Badge variant="secondary" className="w-fit text-[10px]">
                    {msg.type}
                  </Badge>
                )}
              </div>
            ))}
          </div>
        </ScrollArea>
      )}
    </div>
  )
}

function WavesTab() {
  const { data: tasksData, isLoading, error } = useQuery({
    queryKey: ["coordination-tasks"],
    queryFn: listTasks,
  })

  const tasks = tasksData?.tasks ?? []

  const waves = (() => {
    if (tasks.length === 0) return []
    const waveMap = new Map<number, Task[]>()
    for (const t of tasks) {
      const existing = waveMap.get(t.wave) ?? []
      existing.push(t)
      waveMap.set(t.wave, existing)
    }
    return Array.from(waveMap.entries())
      .sort(([a], [b]) => a - b)
      .map(([waveNum, waveTasks]) => {
        const completed = waveTasks.filter(
          (t) => t.status === "completed"
        ).length
        const failed = waveTasks.filter((t) => t.status === "failed").length
        const inProgress = waveTasks.filter(
          (t) => t.status === "in_progress"
        ).length
        const isActive = inProgress > 0 || (completed + failed < waveTasks.length && completed + failed > 0)
        return {
          wave: waveNum,
          tasks: waveTasks.length,
          completed,
          failed,
          inProgress,
          progress: waveTasks.length > 0 ? (completed / waveTasks.length) * 100 : 0,
          isActive,
        }
      })
  })()

  return (
    <>
      {isLoading && (
        <div className="flex items-center justify-center py-12 text-muted-foreground">
          <IconLoader className="size-5 animate-spin mr-2" />
          <span className="text-sm">Loading waves...</span>
        </div>
      )}

      {error && (
        <div className="py-12 text-center text-sm text-destructive">
          Failed to load waves: {error.message}
        </div>
      )}

      {!isLoading && !error && waves.length === 0 && (
        <div className="py-12 text-center text-sm text-muted-foreground">
          No waves to display. Create tasks to see wave execution.
        </div>
      )}

      {!isLoading && !error && waves.length > 0 && (
        <div className="flex flex-col gap-3">
          {waves.map((w) => (
            <Card key={w.wave} size="sm" className={w.isActive ? "ring-cyan-500/40" : ""}>
              <CardContent className="pb-2 pt-0">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">Wave {w.wave}</span>
                    {w.isActive && (
                      <Badge variant="outline" className="bg-cyan-500/15 text-cyan-600 border-cyan-500/25 text-[10px]">
                        Active
                      </Badge>
                    )}
                    {w.progress === 100 && w.failed === 0 && (
                      <Badge variant="outline" className="bg-green-500/15 text-green-600 border-green-500/25 text-[10px]">
                        Complete
                      </Badge>
                    )}
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {w.completed}/{w.tasks} tasks
                  </span>
                </div>
                <Progress
                  value={w.progress}
                  className={`h-2 ${w.isActive ? "[&>div]:bg-cyan-500" : "[&>div]:bg-green-500"}`}
                />
                <div className="flex gap-3 mt-1.5 text-[11px] text-muted-foreground">
                  <span className="text-yellow-500">{w.inProgress} in progress</span>
                  <span className="text-green-500">{w.completed} completed</span>
                  {w.failed > 0 && (
                    <span className="text-red-500">{w.failed} failed</span>
                  )}
                  <span>{w.tasks - w.completed - w.failed - w.inProgress} pending</span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </>
  )
}

export function AgentTeamsPage() {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ["coordination-stats"],
    queryFn: getCoordinationStats,
    refetchInterval: 5000,
  })

  const activeAgents = (stats?.active_agents as number) ?? 0
  const pendingTasks = (stats?.pending as number) ?? 0
  const inProgress = (stats?.in_progress as number) ?? 0
  const completed = (stats?.completed as number) ?? 0

  return (
    <div className="flex h-full flex-col overflow-hidden">
      <PageHeader title="Agent Teams">
        <Button size="sm" variant="outline">
          <IconRobot className="size-4" />
          Manage Agents
        </Button>
      </PageHeader>

      <div className="flex-1 overflow-auto px-6 py-4">
        <CardHeader className="px-0 pt-0">
          <CardTitle className="text-base text-muted-foreground">
            Monitor and manage multi-agent task coordination, inter-agent messages, and wave execution.
          </CardTitle>
        </CardHeader>

        <div className="grid grid-cols-4 gap-3 mb-6">
          <StatCard
            label="Active Agents"
            value={statsLoading ? 0 : activeAgents}
            color="text-cyan-500"
          />
          <StatCard
            label="Pending Tasks"
            value={statsLoading ? 0 : pendingTasks}
            color="text-yellow-500"
          />
          <StatCard
            label="In Progress"
            value={statsLoading ? 0 : inProgress}
            color="text-cyan-600"
          />
          <StatCard
            label="Completed"
            value={statsLoading ? 0 : completed}
            color="text-green-500"
          />
        </div>

        <Tabs defaultValue="tasks">
          <TabsList>
            <TabsTrigger value="tasks">Tasks</TabsTrigger>
            <TabsTrigger value="messages">Messages</TabsTrigger>
            <TabsTrigger value="waves">Waves</TabsTrigger>
          </TabsList>
          <TabsContent value="tasks">
            <TasksTab />
          </TabsContent>
          <TabsContent value="messages">
            <MessagesTab />
          </TabsContent>
          <TabsContent value="waves">
            <WavesTab />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}
