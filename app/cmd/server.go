package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andresbott/dashi/app/metainfo"
	"github.com/andresbott/dashi/app/router"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func serverCmd() *cobra.Command {
	var configFile = "./config.yaml"
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start a web server",
		Long:  "start the dashi web server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(configFile)
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", configFile, "config file")
	return cmd
}

func runServer(configFile string) error {
	// ——— Config and logger ———
	cfg, err := getAppCfg(configFile)
	if err != nil {
		return err
	}

	l, err := defaultLogger(GetLogLevel(cfg.Env.LogLevel))
	if err != nil {
		return err
	}

	l.Info("App startup",
		slog.String("component", "startup"),
		slog.String("version", metainfo.Version),
		slog.String("Build Date", metainfo.BuildTime),
		slog.String("commit", metainfo.ShaVer),
	)
	for _, m := range cfg.Msgs {
		if m.Level == "info" {
			l.Info(m.Msg, slog.String("component", "config"))
		} else {
			l.Debug(m.Msg, slog.String("component", "config"))
		}
	}

	// ——— Start servers ———
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	routerCfg := router.Cfg{
		Ctx:            ctx,
		Logger:         l,
		ProductionMode: cfg.Env.Production,
		DataDir:        cfg.DataDir,
	}

	// Build handlers based on which servers are enabled
	var viewerHandler *router.ViewerHandler
	var editorHandler *router.EditorHandler

	viewerEnabled := cfg.Server.Viewer.Enabled
	editorEnabled := cfg.Server.Editor.Enabled

	switch {
	case viewerEnabled && editorEnabled:
		viewerHandler, editorHandler, err = router.NewBoth(routerCfg)
		if err != nil {
			return fmt.Errorf("failed to create router: %w", err)
		}
	case viewerEnabled:
		l.Info("Editor server disabled")
		viewerHandler, err = router.NewViewer(routerCfg)
		if err != nil {
			return fmt.Errorf("failed to create router: %w", err)
		}
	case editorEnabled:
		l.Info("Viewer server disabled")
		editorHandler, err = router.NewEditor(routerCfg)
		if err != nil {
			return fmt.Errorf("failed to create router: %w", err)
		}
	}

	g, gCtx := errgroup.WithContext(ctx)

	var servers []*http.Server

	if viewerHandler != nil {
		servers = append(servers, startServer(g, gCtx, l, "viewer", cfg.Server.Viewer.Addr(), viewerHandler))
	}
	if editorHandler != nil {
		servers = append(servers, startServer(g, gCtx, l, "editor", cfg.Server.Editor.Addr(), editorHandler))
	}
	if cfg.Obs.Enabled {
		// TODO: add prometheus metrics handler
		servers = append(servers, startServer(g, gCtx, l, "observability", cfg.Obs.Addr(), nil))
	} else {
		l.Info("Observability server disabled")
	}

	// Graceful shutdown
	g.Go(func() error {
		<-gCtx.Done()
		l.Info("Shutting down servers...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		for _, srv := range servers {
			_ = srv.Shutdown(shutdownCtx)
		}
		return nil
	})

	return g.Wait()
}

// startServer creates an HTTP server and launches it in the errgroup.
func startServer(g *errgroup.Group, gCtx context.Context, l *slog.Logger, name, addr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return gCtx
		},
	}
	g.Go(func() error {
		l.Info("Starting "+name+" server", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	return srv
}
