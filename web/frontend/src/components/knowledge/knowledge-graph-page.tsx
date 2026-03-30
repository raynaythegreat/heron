import { IconLoader2, IconNetwork, IconSearch } from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import { useMemo, useState } from "react"

import {
  getGraph,
  getGraphStats,
  searchGraph,
  type GraphNode,
} from "@/api/knowledge-graph"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"

export function KnowledgeGraphPage() {
  const [search, setSearch] = useState("")
  const [searchQuery, setSearchQuery] = useState("")

  const statsQuery = useQuery({
    queryKey: ["knowledge-graph", "stats"],
    queryFn: getGraphStats,
  })

  const graphQuery = useQuery({
    queryKey: ["knowledge-graph"],
    queryFn: getGraph,
  })

  const searchQueryResult = useQuery({
    queryKey: ["knowledge-graph", "search", searchQuery],
    queryFn: () => searchGraph(searchQuery),
    enabled: searchQuery.length > 0,
  })

  const totalNodes = (statsQuery.data?.total_nodes as number) ?? graphQuery.data?.nodes.length ?? 0
  const totalEdges = (statsQuery.data?.total_edges as number) ?? graphQuery.data?.edges.length ?? 0
  const nodeTypes = (statsQuery.data?.node_types as Record<string, number>) ?? {}

  const handleSearch = useMemo(() => {
    let timeout: ReturnType<typeof setTimeout>
    return (value: string) => {
      clearTimeout(timeout)
      setSearch(value)
      timeout = setTimeout(() => {
        setSearchQuery(value.trim())
      }, 300)
    }
  }, [])

  const searchNodes = searchQueryResult.data?.results ?? []

  const statsLoading = statsQuery.isLoading && graphQuery.isLoading

  return (
    <div className="flex flex-1 flex-col overflow-hidden">
      <PageHeader title="Knowledge Graph" />

      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Card className="border-cyan-500/20 bg-gradient-to-br from-cyan-500/5 to-transparent">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Nodes
              </CardTitle>
            </CardHeader>
            <CardContent>
              {statsLoading ? (
                <Skeleton className="h-9 w-16" />
              ) : (
                <span className="text-3xl font-bold text-cyan-400">
                  {totalNodes}
                </span>
              )}
            </CardContent>
          </Card>

          <Card className="border-cyan-500/20 bg-gradient-to-br from-cyan-500/5 to-transparent">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Total Edges
              </CardTitle>
            </CardHeader>
            <CardContent>
              {statsLoading ? (
                <Skeleton className="h-9 w-16" />
              ) : (
                <span className="text-3xl font-bold text-cyan-400">
                  {totalEdges}
                </span>
              )}
            </CardContent>
          </Card>

          <Card className="border-cyan-500/20 bg-gradient-to-br from-cyan-500/5 to-transparent">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Node Types
              </CardTitle>
            </CardHeader>
            <CardContent>
              {statsLoading ? (
                <div className="flex flex-wrap gap-1.5">
                  <Skeleton className="h-5 w-16" />
                  <Skeleton className="h-5 w-12" />
                  <Skeleton className="h-5 w-14" />
                </div>
              ) : Object.keys(nodeTypes).length > 0 ? (
                <div className="flex flex-wrap gap-1.5">
                  {Object.entries(nodeTypes).map(([type, count]) => (
                    <Badge
                      key={type}
                      variant="secondary"
                      className="border-cyan-500/30 bg-cyan-500/10 text-cyan-300 hover:bg-cyan-500/20"
                    >
                      {type}
                      <span className="ml-1 text-cyan-400">{count}</span>
                    </Badge>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">No types yet</p>
              )}
            </CardContent>
          </Card>
        </div>

        <Card className="border-cyan-500/20 bg-gradient-to-br from-cyan-500/5 to-transparent">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <IconNetwork className="size-5 text-cyan-400" />
              Graph Visualization
            </CardTitle>
          </CardHeader>
          <CardContent>
            {graphQuery.isLoading ? (
              <div className="flex min-h-[400px] items-center justify-center rounded-lg border border-dashed border-cyan-500/30 bg-background/50">
                <IconLoader2 className="text-muted-foreground size-6 animate-spin" />
              </div>
            ) : graphQuery.data && graphQuery.data.nodes.length > 0 ? (
              <div className="flex min-h-[400px] items-center justify-center rounded-lg border border-dashed border-cyan-500/30 bg-background/50">
                <p className="text-muted-foreground">
                  Graph visualization will render here
                </p>
              </div>
            ) : (
              <div className="flex min-h-[400px] items-center justify-center rounded-lg border border-dashed border-cyan-500/30 bg-background/50">
                <p className="text-muted-foreground">
                  No graph data available. Run a scan to populate the knowledge graph.
                </p>
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="border-cyan-500/20 bg-gradient-to-br from-cyan-500/5 to-transparent">
          <CardHeader>
            <CardTitle className="text-base">Search Nodes</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="relative">
              <IconSearch className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search by name or type..."
                value={search}
                onChange={(e) => handleSearch(e.target.value)}
                className="border-cyan-500/20 pl-9 focus-visible:ring-cyan-500/50"
              />
            </div>

            {searchQuery.length > 0 && searchQueryResult.isLoading && (
              <div className="flex items-center justify-center py-8">
                <IconLoader2 className="text-muted-foreground size-5 animate-spin" />
              </div>
            )}

            {searchQuery.length > 0 && !searchQueryResult.isLoading && (
              <div className="space-y-2">
                {searchNodes.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No nodes found</p>
                ) : (
                  searchNodes.map((node: GraphNode) => (
                    <div
                      key={node.id}
                      className="flex items-center justify-between rounded-lg border border-border/50 px-3 py-2 transition-colors hover:bg-muted/50"
                    >
                      <span className="text-sm font-medium">{node.name}</span>
                      <Badge
                        variant="secondary"
                        className="border-cyan-500/30 bg-cyan-500/10 text-cyan-300"
                      >
                        {node.type}
                      </Badge>
                    </div>
                  ))
                )}
              </div>
            )}

            {!search.trim() && (
              <p className="text-sm text-muted-foreground">
                Type to search across {totalNodes} nodes in the knowledge
                graph.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
