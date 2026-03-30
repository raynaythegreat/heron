import {
  IconDownload,
  IconLoader2,
  IconSearch,
  IconShieldCheck,
  IconShieldOff,
  IconPlus,
  IconX,
} from "@tabler/icons-react"
import * as React from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

import {
  type Plugin,
  listPlugins,
  searchPlugins,
  togglePlugin,
  scanPlugin,
  installPlugin,
} from "@/api/plugins-api"
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
import { Skeleton } from "@/components/ui/skeleton"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"

type Tier = Plugin["tier"]

const TIER_CONFIG: Record<
  Tier,
  { label: string; color: string; badgeClass: string }
> = {
  system: {
    label: "System",
    color: "text-cyan-400",
    badgeClass:
      "bg-cyan-500/15 text-cyan-400 border-cyan-500/20",
  },
  curated: {
    label: "Curated",
    color: "text-teal-400",
    badgeClass:
      "bg-teal-500/15 text-teal-400 border-teal-500/20",
  },
  experimental: {
    label: "Experimental",
    color: "text-amber-400",
    badgeClass:
      "bg-amber-500/15 text-amber-400 border-amber-500/20",
  },
  local: {
    label: "Local",
    color: "text-cyan-400",
    badgeClass:
      "bg-cyan-500/15 text-cyan-400 border-cyan-500/20",
  },
}

const MOCK_PLUGINS: Plugin[] = [
  {
    id: "terminal-control",
    name: "Terminal Control",
    description:
      "Full terminal access with command history, session management, and safe sandboxed execution.",
    tier: "system",
    enabled: true,
    trust_score: 100,
    installed: true,
    scan_status: "clean",
  },
  {
    id: "file-manager",
    name: "File Manager",
    description:
      "Browse, read, write, and manage files and directories with permission controls.",
    tier: "system",
    enabled: true,
    trust_score: 100,
    installed: true,
    scan_status: "clean",
  },
  {
    id: "web-search",
    name: "Web Search",
    description:
      "Search the web for information, documentation, and real-time data from multiple sources.",
    tier: "system",
    enabled: true,
    trust_score: 98,
    installed: true,
    scan_status: "clean",
  },
  {
    id: "code-review",
    name: "Code Review",
    description:
      "Automated code review with best practices, security checks, and style suggestions.",
    tier: "curated",
    enabled: true,
    trust_score: 95,
    installed: true,
    scan_status: "clean",
  },
  {
    id: "test-generator",
    name: "Test Generator",
    description:
      "Generate comprehensive unit and integration tests from source code and specifications.",
    tier: "curated",
    enabled: false,
    trust_score: 92,
    installed: true,
    scan_status: "clean",
  },
  {
    id: "documentation-writer",
    name: "Documentation Writer",
    description:
      "Generate and maintain project documentation, API references, and inline code docs.",
    tier: "curated",
    enabled: true,
    trust_score: 90,
    installed: true,
    scan_status: "clean",
  },
  {
    id: "auto-deploy",
    name: "Auto-Deploy",
    description:
      "Automate deployment pipelines with rollback support and environment management.",
    tier: "experimental",
    enabled: false,
    trust_score: 72,
    installed: true,
    scan_status: "warning",
  },
  {
    id: "performance-analyzer",
    name: "Performance Analyzer",
    description:
      "Profile and analyze application performance with bottleneck detection and optimization hints.",
    tier: "experimental",
    enabled: false,
    trust_score: 68,
    installed: true,
    scan_status: "unscanned",
  },
]

function PluginCardSkeleton() {
  return (
    <Card size="sm" className="flex flex-col">
      <CardHeader className="border-b border-border/40 pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1 space-y-2">
            <Skeleton className="h-5 w-28" />
            <Skeleton className="h-5 w-16" />
          </div>
          <Skeleton className="h-5 w-9 rounded-full" />
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col justify-between gap-3 pt-3">
        <div className="space-y-1.5">
          <Skeleton className="h-3 w-full" />
          <Skeleton className="h-3 w-3/4" />
          <Skeleton className="h-3 w-1/2" />
        </div>
        <div className="flex items-center justify-between gap-2">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-5 w-5 rounded" />
        </div>
      </CardContent>
    </Card>
  )
}

function TrustScoreBar({ score }: { score: number }) {
  const barColor =
    score >= 90
      ? "bg-teal-500"
      : score >= 70
        ? "bg-amber-500"
        : "bg-red-500"

  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 w-16 overflow-hidden rounded-full bg-muted">
        <div
          className={`h-full rounded-full transition-all ${barColor}`}
          style={{ width: `${score}%` }}
        />
      </div>
      <span className="text-muted-foreground text-xs font-medium">{score}%</span>
    </div>
  )
}

function ScanStatusIcon({ status }: { status: Plugin["scan_status"] }) {
  switch (status) {
    case "clean":
      return (
        <span className="flex items-center gap-1 text-teal-400 text-xs">
          <IconShieldCheck className="size-3.5" />
          Clean
        </span>
      )
    case "scanning":
      return (
        <span className="flex items-center gap-1 text-cyan-400 text-xs animate-pulse">
          <IconShieldCheck className="size-3.5 animate-spin" />
          Scanning...
        </span>
      )
    case "warning":
      return (
        <span className="flex items-center gap-1 text-amber-400 text-xs">
          <IconShieldOff className="size-3.5" />
          Warnings
        </span>
      )
    case "unscanned":
      return (
        <span className="flex items-center gap-1 text-muted-foreground text-xs">
          <IconShieldOff className="size-3.5" />
          Unscanned
        </span>
      )
  }
}

function InstallSheet({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const [url, setUrl] = React.useState("")
  const queryClient = useQueryClient()

  const installMutation = useMutation({
    mutationFn: (pluginUrl: string) => installPlugin(pluginUrl),
    onSuccess: () => {
      toast.success("Plugin installed successfully")
      setUrl("")
      onOpenChange(false)
      void queryClient.invalidateQueries({ queryKey: ["plugins"] })
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : "Failed to install plugin")
    },
  })

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>Install Plugin</SheetTitle>
          <SheetDescription>
            Enter a plugin source URL or registry path to install.
          </SheetDescription>
        </SheetHeader>
        <div className="mt-6 space-y-4">
          <Input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://github.com/user/plugin or registry://plugin-name"
            onKeyDown={(e) => {
              if (e.key === "Enter" && url.trim()) {
                installMutation.mutate(url.trim())
              }
            }}
          />
          <Button
            className="w-full bg-teal-600 hover:bg-teal-700"
            onClick={() => installMutation.mutate(url.trim())}
            disabled={!url.trim() || installMutation.isPending}
          >
            {installMutation.isPending ? (
              <IconLoader2 className="mr-2 size-4 animate-spin" />
            ) : (
              <IconDownload className="mr-2 size-4" />
            )}
            Install Plugin
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  )
}

function PluginCard({
  plugin,
  onToggle,
  onScan,
  togglePending,
  scanPending,
}: {
  plugin: Plugin
  onToggle: (id: string, enabled: boolean) => void
  onScan: (id: string) => void
  togglePending: boolean
  scanPending: boolean
}) {
  const tier = TIER_CONFIG[plugin.tier]

  return (
    <Card size="sm" className="flex flex-col">
      <CardHeader className="border-b border-border/40 pb-3">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <CardTitle className="truncate">{plugin.name}</CardTitle>
            <Badge
              variant="outline"
              className={`mt-1.5 text-xs ${tier.badgeClass}`}
            >
              {tier.label}
            </Badge>
          </div>
          <Switch
            checked={plugin.enabled}
            onCheckedChange={(checked) => onToggle(plugin.id, checked)}
            disabled={togglePending}
            size="sm"
          />
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col justify-between gap-3 pt-3">
        <p className="text-muted-foreground line-clamp-3 text-xs">
          {plugin.description}
        </p>
        <div className="flex items-center justify-between gap-2">
          <div className="flex items-center gap-3">
            <TrustScoreBar score={plugin.trust_score} />
            <ScanStatusIcon
              status={scanPending ? "scanning" : plugin.scan_status}
            />
          </div>
          <Button
            variant="ghost"
            size="icon-xs"
            className="text-muted-foreground hover:text-teal-400 hover:bg-teal-500/10 ml-auto shrink-0"
            onClick={() => onScan(plugin.id)}
            disabled={scanPending}
            title="Security scan"
          >
            {scanPending ? (
              <IconLoader2 className="size-3.5 animate-spin" />
            ) : (
              <IconShieldCheck className="size-3.5" />
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export function PluginsPage() {
  const queryClient = useQueryClient()
  const [search, setSearch] = React.useState("")
  const [activeTab, setActiveTab] = React.useState<Tier>("system")
  const [installOpen, setInstallOpen] = React.useState(false)
  const [searchDebounce, setSearchDebounce] = React.useState("")

  React.useEffect(() => {
    const timer = setTimeout(() => setSearchDebounce(search), 300)
    return () => clearTimeout(timer)
  }, [search])

  const isSearching = searchDebounce.length > 0

  const { data: listData, isLoading: listLoading } = useQuery({
    queryKey: ["plugins", "list"],
    queryFn: () => listPlugins(),
    enabled: !isSearching,
  })

  const { data: searchData, isLoading: searchLoading } = useQuery({
    queryKey: ["plugins", "search", searchDebounce],
    queryFn: () => searchPlugins(searchDebounce),
    enabled: isSearching,
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      togglePlugin(id, enabled),
    onSuccess: (_, variables) => {
      toast.success(variables.enabled ? "Plugin enabled" : "Plugin disabled")
      void queryClient.invalidateQueries({ queryKey: ["plugins"] })
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : "Failed to toggle plugin")
    },
  })

  const scanMutation = useMutation({
    mutationFn: (id: string) => scanPlugin(id),
    onSuccess: (data) => {
      if (data.security && data.security.score === 100) {
        toast.success("Security scan complete - no issues found")
      } else if (data.security && data.security.score < 100) {
        toast.warning("Security scan found warnings")
      }
      void queryClient.invalidateQueries({ queryKey: ["plugins"] })
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : "Scan failed")
    },
  })

  const isLoading = isSearching ? searchLoading : listLoading

  const plugins: Plugin[] = React.useMemo(() => {
    const source = isSearching ? searchData?.results : listData?.plugins
    if (source && source.length > 0) return source
    return MOCK_PLUGINS
  }, [isSearching, listData, searchData])

  const filteredPlugins = plugins.filter((p) => {
    if (p.tier !== activeTab) return false
    return true
  })

  const totalPlugins = plugins.length
  const enabledCount = plugins.filter((p) => p.enabled).length
  const cleanCount = plugins.filter((p) => p.scan_status === "clean").length
  const allScanned = plugins.every((p) => p.scan_status !== "unscanned")

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title="Plugin Marketplace"
        children={
          <Button
            className="bg-teal-600 hover:bg-teal-700"
            onClick={() => setInstallOpen(true)}
          >
            <IconPlus className="size-4" />
            Install Plugin
          </Button>
        }
      />

      <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-4 md:p-6">
        <div className="flex shrink-0 flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="relative flex-1 sm:max-w-xs">
            <IconSearch className="text-muted-foreground absolute left-3 top-1/2 size-4 -translate-y-1/2" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search plugins..."
              className="pl-9"
            />
          </div>
          <div className="flex items-center gap-4 text-xs text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <span className="text-foreground font-medium">{totalPlugins}</span>
              Total
            </span>
            <span className="flex items-center gap-1.5">
              <span className="text-teal-400 font-medium">{enabledCount}</span>
              Enabled
            </span>
            <span className="flex items-center gap-1.5">
              <IconShieldCheck
                className={`size-3.5 ${allScanned ? "text-teal-400" : "text-amber-400"}`}
              />
              <span className={allScanned ? "text-teal-400" : "text-amber-400"}>
                {allScanned ? "All Scanned" : `${cleanCount}/${totalPlugins} Scanned`}
              </span>
            </span>
          </div>
        </div>

        <Tabs
          value={activeTab}
          onValueChange={(v) => setActiveTab(v as Tier)}
          className="shrink-0"
        >
          <TabsList>
            <TabsTrigger value="system">
              <span className="text-cyan-400">System</span>
            </TabsTrigger>
            <TabsTrigger value="curated">
              <span className="text-teal-400">Curated</span>
            </TabsTrigger>
            <TabsTrigger value="experimental">
              <span className="text-amber-400">Experimental</span>
            </TabsTrigger>
            <TabsTrigger value="local">
              <span className="text-cyan-400">Local</span>
            </TabsTrigger>
          </TabsList>

          {(["system", "curated", "experimental", "local"] as Tier[]).map(
            (tier) => (
              <TabsContent key={tier} value={tier}>
                {isLoading ? (
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {Array.from({ length: 3 }).map((_, i) => (
                      <PluginCardSkeleton key={i} />
                    ))}
                  </div>
                ) : filteredPlugins.length === 0 ? (
                  <div className="flex flex-col items-center justify-center gap-3 py-12 text-center">
                    <IconX className="text-muted-foreground/40 size-10" />
                    <p className="text-muted-foreground text-sm">
                      {searchDebounce
                        ? "No plugins match your search"
                        : `No ${TIER_CONFIG[tier].label.toLowerCase()} plugins available`}
                    </p>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {filteredPlugins.map((plugin) => (
                      <PluginCard
                        key={plugin.id}
                        plugin={plugin}
                        onToggle={(id, enabled) =>
                          toggleMutation.mutate({ id, enabled })
                        }
                        onScan={(id) => scanMutation.mutate(id)}
                        togglePending={
                          toggleMutation.isPending &&
                          toggleMutation.variables?.id === plugin.id
                        }
                        scanPending={
                          scanMutation.isPending &&
                          scanMutation.variables === plugin.id
                        }
                      />
                    ))}
                  </div>
                )}
              </TabsContent>
            ),
          )}
        </Tabs>
      </div>

      <InstallSheet open={installOpen} onOpenChange={setInstallOpen} />
    </div>
  )
}
