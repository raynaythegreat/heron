package web

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/raynaythegreat/heron/cmd/heron/internal/run"
	webconsole "github.com/raynaythegreat/heron/web/backend"
	"github.com/raynaythegreat/heron/web/backend/utils"
)

func NewWebCommand() *cobra.Command {
	var opts webconsole.Options

	cmd := &cobra.Command{
		Use:   "web",
		Short: "Manage the web console",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to start if no subcommand
			return startWeb(cmd, args, opts)
		},
	}

	startCmd := &cobra.Command{
		Use:   "start [config.json]",
		Short: "Start the web console in background",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startWeb(cmd, args, opts)
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the web console",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := run.ReadPID("web")
			if err != nil {
				return fmt.Errorf("could not find PID for web: %v", err)
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Errorf("could not find process: %v", err)
			}
			err = process.Signal(syscall.SIGTERM)
			if err != nil {
				return fmt.Errorf("could not signal process: %v", err)
			}
			run.RemovePID("web")
			fmt.Println("Web console stopped.")
			return nil
		},
	}

	cmd.AddCommand(startCmd, stopCmd)

	// Flags for start command
	for _, c := range []*cobra.Command{cmd, startCmd} {
		c.Flags().StringVar(&opts.Port, "port", "18800", "Port to listen on")
		c.Flags().BoolVar(&opts.Public, "public", false, "Listen on all interfaces (0.0.0.0)")
		c.Flags().BoolVar(&opts.NoBrowser, "no-browser", false, "Do not auto-open browser on startup")
		c.Flags().StringVar(&opts.Lang, "lang", "", "Language: en (English) or zh (Chinese)")
		c.Flags().BoolVar(&opts.Console, "console", false, "Console mode, no system tray GUI")
	}

	return cmd
}

func startWeb(cmd *cobra.Command, args []string, opts webconsole.Options) error {
	if len(args) > 0 {
		opts.ConfigPath = args[0]
	} else {
		opts.ConfigPath = utils.GetDefaultConfigPath()
	}
	opts.ExplicitPort = cmd.Flags().Changed("port")
	opts.ExplicitPublic = cmd.Flags().Changed("public")

	// Start in background
	executable, _ := os.Executable()
	argsToPass := append([]string{"web", "start", "--console"}, os.Args[2:]...)
	proc := exec.Command(executable, argsToPass...)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := proc.Start(); err != nil {
		return err
	}

	run.WritePID("web", proc.Process.Pid)
	fmt.Printf("Web console started in background (PID: %d)\n", proc.Process.Pid)
	return nil
}
