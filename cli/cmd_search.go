package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	var year int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search NeurIPS papers by title",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			if query == "" {
				return codeError(exitUsage, fmt.Errorf("query cannot be empty"))
			}
			limit := a.effectiveLimit(0)
			papers, err := a.client.Search(cmd.Context(), query, year, limit)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(papers, len(papers))
		},
	}
	cmd.Flags().IntVar(&year, "year", 2024, "NeurIPS proceedings year")
	return cmd
}
