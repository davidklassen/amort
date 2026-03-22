package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/davidklassen/amort/store"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show full proposal plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get(daemonURL("/api/proposals/" + args[0]))
			if err != nil {
				return fmt.Errorf("is amort running? %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == 404 {
				return fmt.Errorf("proposal %s not found", args[0])
			}

			var p store.Proposal
			if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
				return err
			}

			fmt.Printf("# %s\n\nStatus: %s\n\n%s\n\n---\n\n%s\n", p.Title, p.Status, p.Summary, p.Plan)
			return nil
		},
	}
}
