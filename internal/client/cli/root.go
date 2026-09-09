package cli

import (
	"github.com/michaeltukdev/Potok/internal/client/cli/commands"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "potok",
		Short: "Encrypted backup and sync for Obsidian vaults",
		Long: "Potok backs up and syncs Obsidian vaults to a server you control.\n" +
			"Vaults are encrypted on this device before upload, so the server\n" +
			"never sees your passphrase or your notes.",

		SilenceUsage:  true,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	root.AddCommand(
		commands.NewInitCmd(),
		commands.NewVaultAddCmd(),
		commands.NewVaultRemoveCmd(),
		commands.NewDoctorCmd(),
		commands.NewRemoteListCmd(),
		commands.NewRemoteDeleteCmd(),
	)

	return root
}

func Execute() error {
	return newRootCmd().Execute()
}
