import { MODEL_CATALOG } from "@/components/models/catalog-data"

export interface ModelCatalogEntry {
  id: string
  provider: string
  providerKey: string
  name: string
  category: string
  desc: string
  capabilities?: string[]
  contextLength?: number
  pricing?: { input: number; output: number }
}

export async function fetchModelCatalog(): Promise<ModelCatalogEntry[]> {
  try {
    const response = await fetch("/model-catalog.json")
    if (response.ok) {
      return await response.json()
    }
  } catch {
    // fall through to hardcoded fallback
  }
  return getDefaultCatalog()
}

function getDefaultCatalog(): ModelCatalogEntry[] {
  return MODEL_CATALOG.flatMap((provider) =>
    provider.models.map((model) => ({
      id: model.id,
      provider: provider.provider,
      providerKey: provider.providerKey,
      name: model.name,
      category: provider.category,
      desc: model.desc,
    })),
  )
}
