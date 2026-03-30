import { createFileRoute } from "@tanstack/react-router"

import { KnowledgeGraphPage } from "@/components/knowledge/knowledge-graph-page"

export const Route = createFileRoute("/knowledge")({
  component: KnowledgeGraphPage,
})
