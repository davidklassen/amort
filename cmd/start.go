package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/davidklassen/amort/explore"
	"github.com/davidklassen/amort/server"
	"github.com/davidklassen/amort/store"
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	var qcap int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the amort daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := ".amort"
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			s, err := store.Open(filepath.Join(dir, "amort.db"))
			if err != nil {
				return err
			}
			defer func() { _ = s.Close() }()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			loop := explore.NewLoop(s, qcap)
			go loop.Run(ctx)

			srv := server.New(s, loop)
			httpServer := &http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: srv.Handler(),
			}

			go func() {
				<-ctx.Done()
				_ = httpServer.Close()
			}()

			slog.Info("amort started", "url", fmt.Sprintf("http://localhost:%d", port))
			if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
				return err
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&qcap, "cap", 10, "max pending proposals")

	return cmd
}
