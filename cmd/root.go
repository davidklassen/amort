package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var port int

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "amort",
		Short:         "Continuous code improvement agent",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().IntVar(&port, "port", 4444, "daemon port")

	cmd.AddCommand(newStartCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newApproveCmd())
	cmd.AddCommand(newRejectCmd())
	cmd.AddCommand(newResumeCmd())

	return cmd
}

func Execute() error {
	return newRootCmd().Execute()
}

func daemonURL(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", port, path)
}
