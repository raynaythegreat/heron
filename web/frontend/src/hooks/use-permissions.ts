import { atom, useAtom } from "jotai"

interface HeronPermissions {
  terminal: boolean
  filesystem: boolean
  filesystemPaths: string[]
  microphone: boolean
  network: boolean
  browser: boolean
  setupCompleted: boolean
}

const DEFAULT_PERMISSIONS: HeronPermissions = {
  terminal: false,
  filesystem: false,
  filesystemPaths: [],
  microphone: false,
  network: true,
  browser: false,
  setupCompleted: false,
}

function loadPermissions(): HeronPermissions {
  try {
    const stored = localStorage.getItem("heron:permissions")
    if (stored) return { ...DEFAULT_PERMISSIONS, ...JSON.parse(stored) }
  } catch {
    // ignore
  }
  return DEFAULT_PERMISSIONS
}

function savePermissions(perms: HeronPermissions) {
  localStorage.setItem("heron:permissions", JSON.stringify(perms))
}

export const permissionsAtom = atom<HeronPermissions>(loadPermissions())

export function usePermissions() {
  const [permissions, setPermissions] = useAtom(permissionsAtom)

  const updatePermission = (key: keyof HeronPermissions, value: unknown) => {
    const updated = { ...permissions, [key]: value }
    setPermissions(updated)
    savePermissions(updated)
  }

  const completeSetup = () => {
    const updated = { ...permissions, setupCompleted: true }
    setPermissions(updated)
    savePermissions(updated)
  }

  const resetPermissions = () => {
    setPermissions(DEFAULT_PERMISSIONS)
    savePermissions(DEFAULT_PERMISSIONS)
  }

  return { permissions, updatePermission, completeSetup, resetPermissions }
}
