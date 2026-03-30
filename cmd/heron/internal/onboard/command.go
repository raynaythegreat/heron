package onboard

import (
	"embed"

	"github.com/spf13/cobra"
)

//go:generate cp -r ../../../../workspace .
//go:embed workspace
var embeddedFiles embed.FS

//go:generate cp -r ../../../../skills .
//go:embed skills
var embeddedSkills embed.FS

func NewOnboardCommand() *cobra.Command {
	var encrypt bool
	var legacy bool

	cmd := &cobra.Command{
		Use:     "onboard",
		Aliases: []string{"o"},
		Short:   "Initialize heron configuration and workspace",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				if legacy {
					onboardLegacy(encrypt)
				} else {
					onboard(encrypt)
				}
			} else {
				_ = cmd.Help()
			}
		},
	}

	cmd.Flags().BoolVar(&encrypt, "enc", false,
		"Enable credential encryption (generates SSH key and prompts for passphrase)")
	cmd.Flags().BoolVar(&legacy, "legacy", false,
		"Use the original text-based wizard instead of the TUI wizard")

	return cmd
}
