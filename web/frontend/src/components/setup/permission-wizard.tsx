import {
  IconBrowser,
  IconCheck,
  IconChevronLeft,
  IconChevronRight,
  IconDeviceDesktop,
  IconFolder,
  IconMicrophone,
  IconNetwork,
  IconSparkles,
  IconTerminal2,
} from "@tabler/icons-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"
import { Switch } from "@/components/ui/switch"
import { usePermissions } from "@/hooks/use-permissions"

interface WizardStep {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  content: React.ReactNode
}

const TOTAL_STEPS = 7

export function PermissionWizard() {
  const [currentStep, setCurrentStep] = useState(0)
  const [filesystemPaths, setFilesystemPaths] = useState("/home")
  const { permissions, updatePermission, completeSetup } = usePermissions()

  const handleNext = () => {
    if (currentStep < TOTAL_STEPS - 1) {
      setCurrentStep((s) => s + 1)
    }
  }

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep((s) => s - 1)
    }
  }

  const handleFinish = () => {
    completeSetup()
  }

  const progressPercent = ((currentStep + 1) / TOTAL_STEPS) * 100

  const steps: WizardStep[] = [
    {
      id: "welcome",
      title: "Welcome to Heron",
      description: "Get started by configuring your permissions",
      icon: <IconSparkles className="size-10 text-cyan-400" />,
      content: (
        <div className="space-y-4 text-center">
          <p className="text-muted-foreground">
            Heron needs certain permissions to function properly. You can always change these later in Settings.
          </p>
          <p className="text-sm text-muted-foreground">
            This setup will guide you through granting or denying access to system resources.
          </p>
        </div>
      ),
    },
    {
      id: "terminal",
      title: "Terminal Access",
      description: "Allow shell command execution",
      icon: <IconTerminal2 className="size-10 text-cyan-400" />,
      content: (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Terminal access allows Heron to execute shell commands on your system.
          </p>
          <div className="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-sm text-amber-400">
            Warning: Granting terminal access may pose security risks. Only enable this if you trust the agent's actions.
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Allow terminal access</span>
            <Switch
              checked={permissions.terminal}
              onCheckedChange={(v) => updatePermission("terminal", v)}
            />
          </div>
        </div>
      ),
    },
    {
      id: "filesystem",
      title: "File System",
      description: "Read and write file permissions",
      icon: <IconFolder className="size-10 text-cyan-400" />,
      content: (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Control which directories Heron can read from and write to.
          </p>
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Allow file system access</span>
            <Switch
              checked={permissions.filesystem}
              onCheckedChange={(v) => updatePermission("filesystem", v)}
            />
          </div>
          {permissions.filesystem && (
            <div className="space-y-2">
              <label className="text-sm font-medium">Allowed directories</label>
              <input
                type="text"
                value={filesystemPaths}
                onChange={(e) => {
                  setFilesystemPaths(e.target.value)
                  updatePermission(
                    "filesystemPaths",
                    e.target.value.split(",").map((s) => s.trim()).filter(Boolean),
                  )
                }}
                placeholder="/home,/tmp"
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
              <p className="text-xs text-muted-foreground">Comma-separated paths</p>
            </div>
          )}
        </div>
      ),
    },
    {
      id: "microphone",
      title: "Microphone",
      description: "Voice input permission",
      icon: <IconMicrophone className="size-10 text-cyan-400" />,
      content: (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Allow Heron to access your microphone for voice input in chat.
          </p>
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Allow microphone access</span>
            <Switch
              checked={permissions.microphone}
              onCheckedChange={(v) => updatePermission("microphone", v)}
            />
          </div>
        </div>
      ),
    },
    {
      id: "network",
      title: "Network",
      description: "External API and web access",
      icon: <IconNetwork className="size-10 text-cyan-400" />,
      content: (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Allow Heron to make requests to external APIs and web services.
          </p>
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Allow network access</span>
            <Switch
              checked={permissions.network}
              onCheckedChange={(v) => updatePermission("network", v)}
            />
          </div>
        </div>
      ),
    },
    {
      id: "browser",
      title: "Browser",
      description: "Web browser automation",
      icon: <IconBrowser className="size-10 text-cyan-400" />,
      content: (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Allow Heron to automate web browser actions for testing and scraping.
          </p>
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">Allow browser automation</span>
            <Switch
              checked={permissions.browser}
              onCheckedChange={(v) => updatePermission("browser", v)}
            />
          </div>
        </div>
      ),
    },
    {
      id: "complete",
      title: "Setup Complete",
      description: "Your permissions are configured",
      icon: <IconDeviceDesktop className="size-10 text-cyan-400" />,
      content: (
        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Here's a summary of your permission choices:
          </p>
          <div className="space-y-2">
            {([
              ["Terminal", permissions.terminal],
              ["File System", permissions.filesystem],
              ["Microphone", permissions.microphone],
              ["Network", permissions.network],
              ["Browser", permissions.browser],
            ] as const).map(([label, granted]) => (
              <div key={label} className="flex items-center justify-between text-sm">
                <span>{label}</span>
                {granted ? (
                  <span className="flex items-center gap-1 text-emerald-400">
                    <IconCheck className="size-4" /> Allowed
                  </span>
                ) : (
                  <span className="text-muted-foreground">Denied</span>
                )}
              </div>
            ))}
          </div>
          <p className="text-xs text-muted-foreground">
            You can change these permissions at any time in Settings.
          </p>
        </div>
      ),
    },
  ]

  const step = steps[currentStep]
  const isFirst = currentStep === 0
  const isLast = currentStep === TOTAL_STEPS - 1

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <div className="mb-2 flex justify-center">{step.icon}</div>
          <CardTitle className="text-center">{step.title}</CardTitle>
          <CardDescription className="text-center">{step.description}</CardDescription>
        </CardHeader>
        <CardContent>
          <Progress value={progressPercent} className="mb-6" />
          {step.content}
        </CardContent>
        <CardFooter className="justify-between">
          <Button
            variant="outline"
            size="sm"
            onClick={handlePrev}
            disabled={isFirst}
          >
            <IconChevronLeft className="size-4" />
            Back
          </Button>
          {isLast ? (
            <Button size="sm" onClick={handleFinish}>
              Get Started
            </Button>
          ) : (
            <Button size="sm" onClick={handleNext}>
              Next
              <IconChevronRight className="size-4" />
            </Button>
          )}
        </CardFooter>
      </Card>
    </div>
  )
}
