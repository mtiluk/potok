package commands

import (
	"github.com/spf13/cobra"
)

func NewVaultAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault-add <name>",
		Short: "Register a local folder as a vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// // 1. Validate the name. config.ValidateVaultName already does this.
			// if err := config.ValidateVaultName(name); err != nil {
			// 	return err
			// }

			// // 2. Resolve the path to absolute and check it is a readable directory.
			// //    Warn (don't fail) if there is no .obsidian folder.

			// // 3. Load the config. config.ErrNotFound means run `potok init` first.
			// cfg, err := config.Load()
			// if errors.Is(err, config.ErrNotFound) {
			// 	return errors.New("not initialised, run `potok init` first")
			// }
			// if err != nil {
			// 	return err
			// }

			// // 4. Reject a name that is already registered, before prompting for
			// //    anything — nobody wants to type a passphrase twice and then be told.

			// // 5. Prompt for the passphrase, twice, no echo. Reject empty; reject
			// //    mismatched.

			// // 6. Store the passphrase in the keyring under secrets.VaultKey(name).

			// // 7. Add the vault to the config and save. If this fails, delete the
			// //    keyring entry so a retry isn't blocked by an orphaned passphrase.

			return nil
		},
	}
}
