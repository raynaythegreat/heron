package channel

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/raynaythegreat/heron/cmd/heron/internal/run"
)

func NewChannelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Manage communication channels (gateways)",
	}

	startCmd := &cobra.Command{
		Use:   "start [name]",
		Short: "Start a channel gateway in background",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			executable, _ := os.Executable()
			// Assuming 'heron gateway' manages channels internally
			proc := exec.Command(executable, "gateway", "--channel", name) // Hypothetical gateway command arguments
			proc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := proc.Start(); err != nil {
				return err
			}
			run.WritePID("channel-"+name, proc.Process.Pid)
			fmt.Printf("Channel %s started in background (PID: %d)\n", name, proc.Process.Pid)
			return nil
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop [name]",
		Short: "Stop a channel gateway",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pid, err := run.ReadPID("channel-" + name)
			if err != nil {
				return fmt.Errorf("could not find PID for channel %s: %v", name, err)
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Errorf("could not find process: %v", err)
			}
			process.Signal(syscall.SIGTERM)
			run.RemovePID("channel-" + name)
			fmt.Printf("Channel %s stopped.\n", name)
			return nil
		},
	}

	cmd.AddCommand(startCmd, stopCmd)
	return cmd
}
