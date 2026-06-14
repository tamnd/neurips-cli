package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) listCmd() *cobra.Command {
	var year int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List NeurIPS papers for a given year",
		RunE: func(cmd *cobra.Command, _ []string) error {
			limit := a.effectiveLimit(0)
			papers, err := a.client.List(cmd.Context(), year, limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(papers, len(papers))
		},
	}
	cmd.Flags().IntVar(&year, "year", 2024, "NeurIPS proceedings year")
	return cmd
}
