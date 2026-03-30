// Heron - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 Heron contributors

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/raynaythegreat/heron/cmd/heron/internal"
	"github.com/raynaythegreat/heron/cmd/heron/internal/agent"
	"github.com/raynaythegreat/heron/cmd/heron/internal/auth"
	"github.com/raynaythegreat/heron/cmd/heron/internal/channel"
	"github.com/raynaythegreat/heron/cmd/heron/internal/cron"
	"github.com/raynaythegreat/heron/cmd/heron/internal/gateway"
	"github.com/raynaythegreat/heron/cmd/heron/internal/loop"
	"github.com/raynaythegreat/heron/cmd/heron/internal/migrate"
	"github.com/raynaythegreat/heron/cmd/heron/internal/model"
	"github.com/raynaythegreat/heron/cmd/heron/internal/models"
	"github.com/raynaythegreat/heron/cmd/heron/internal/onboard"
	"github.com/raynaythegreat/heron/cmd/heron/internal/skills"
	"github.com/raynaythegreat/heron/cmd/heron/internal/status"
	"github.com/raynaythegreat/heron/cmd/heron/internal/tui"
	"github.com/raynaythegreat/heron/cmd/heron/internal/version"
	"github.com/raynaythegreat/heron/cmd/heron/internal/web"
	"github.com/raynaythegreat/heron/pkg/config"
)

func NewPicoclawCommand() *cobra.Command {
	short := fmt.Sprintf("%s Heron - Personal AI Assistant v%s\n\n", internal.Logo, config.GetVersion())

	cmd := &cobra.Command{
		Use:     "heron",
		Short:   short,
		Example: "heron version",
	}

	agentCmd := agent.NewAgentCommand()

	// agents is an alias for agent
	agentsCmd := agent.NewAgentCommand()
	agentsCmd.Use = "agents"
	agentsCmd.Short = "Interactive AI chat (alias for 'agent')"
	agentsCmd.Long = agentsCmd.Short

	cmd.AddCommand(
		onboard.NewOnboardCommand(),
		agentCmd,
		agentsCmd,
		auth.NewAuthCommand(),
		channel.NewChannelCommand(),
		gateway.NewGatewayCommand(),
		status.NewStatusCommand(),
		cron.NewCronCommand(),
		loop.NewLoopCommand(),
		migrate.NewMigrateCommand(),
		skills.NewSkillsCommand(),
		model.NewModelCommand(),
		models.NewModelsCommand(),
		web.NewWebCommand(),
		tui.NewTUICommand(),
		version.NewVersionCommand(),
	)

	return cmd
}

const (
	colorPurple = "\033[1;38;2;168;85;247m"
	banner      = "\r\n" +
		colorPurple + " ██████╗  ██████╗████████╗ █████╗ ██╗\n" +
		colorPurple + "██╔═══██╗██╔════╝╚══██╔══╝██╔══██╗██║\n" +
		colorPurple + "██║   ██║██║        ██║   ███████║██║\n" +
		colorPurple + "╚██████╔╝╚██████╗   ██║   ██╔══██║██║\n" +
		colorPurple + " ╚═════╝  ╚═════╝   ╚═╝   ╚═╝  ╚═╝╚═╝\n" +
		"\033[0m\r\n"
)

func main() {
	// Suppress banner for interactive chat commands
	suppressBanner := len(os.Args) >= 2 && (os.Args[1] == "agent" || os.Args[1] == "agents" || os.Args[1] == "tui")
	if !suppressBanner {
		fmt.Printf("%s", banner)
	}
	cmd := NewPicoclawCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
