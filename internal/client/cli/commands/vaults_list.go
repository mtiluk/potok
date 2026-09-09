package commands

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/michaeltukdev/Potok/internal/client/config"
	"github.com/spf13/cobra"
)

func NewVaultsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vaults-list",
		Short: "List vaults registered locally",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if errors.Is(err, config.ErrNotFound) {
				return errors.New(color.RedString("Not initialised, run `potok init` first"))
			}
			if err != nil {
				return err
			}

			if len(cfg.Vaults) == 0 {
				fmt.Println("No vaults registered. Run `potok vault-add <name>` to register one.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPATH\tLAST SYNCED")
			for _, vault := range cfg.Vaults {
				fmt.Fprintf(w, "%s\t%s\t%s\n", vault.Name, vault.Path, lastSyncedString(vault.LastSyncedAt))
			}
			return w.Flush()
		},
	}
}

func lastSyncedString(t *time.Time) string {
	if t == nil {
		return "never"
	}
	return t.Local().Format(time.RFC3339)
}
