import { createFileRoute } from "@tanstack/react-router"

import { HooksPage } from "@/components/config/hooks-page"

export const Route = createFileRoute("/hooks")({
  component: HooksPage,
})
