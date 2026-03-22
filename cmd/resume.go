package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"github.com/davidklassen/amort/store"
	"github.com/spf13/cobra"
)

func newResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <id>",
		Short: "Resume the planning session for a proposal",
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

			claude, err := exec.LookPath("claude")
			if err != nil {
				return fmt.Errorf("claude not found in PATH")
			}

			return syscall.Exec(claude, []string{"claude", "--resume", p.SessionID}, os.Environ())
		},
	}
}
