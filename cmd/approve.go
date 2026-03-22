package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func newApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <id>",
		Short: "Approve a proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := http.Post(daemonURL("/api/proposals/"+args[0]+"/approve"), "", nil)
			if err != nil {
				return fmt.Errorf("is amort running? %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode == 404 {
				return fmt.Errorf("proposal %s not found or already resolved", args[0])
			}

			fmt.Printf("approved %s\n", args[0])
			return nil
		},
	}
}
