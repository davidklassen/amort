package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/davidklassen/amort/store"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List pending proposals",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Get(daemonURL("/api/proposals"))
			if err != nil {
				return fmt.Errorf("is amort running? %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			var proposals []store.Proposal
			if err := json.NewDecoder(resp.Body).Decode(&proposals); err != nil {
				return err
			}

			if len(proposals) == 0 {
				fmt.Println("no pending proposals")
				return nil
			}

			for _, p := range proposals {
				plan := p.Plan
				if len(plan) > 200 {
					plan = plan[:200] + "..."
				}
				fmt.Printf("[%s] %s\n      %s\n      %s\n\n", p.ID, p.Title, p.Summary, plan)
			}
			return nil
		},
	}
}
