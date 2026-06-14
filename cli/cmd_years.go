package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/neurips-cli/neurips"
)

func (a *App) yearsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "years",
		Short: "List available NeurIPS proceedings years",
		RunE: func(cmd *cobra.Command, _ []string) error {
			years := neurips.Years()
			return a.render(years)
		},
	}
}
