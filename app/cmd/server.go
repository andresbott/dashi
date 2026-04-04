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

	// ——— Build the main app handler ———
	appHandler, err := router.New(router.Cfg{
		Logger:         l,
		ProductionMode: cfg.Env.Production,
		DataDir:        cfg.DataDir,
	})
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	// ——— Start servers ———
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)

	// Main server
	mainSrv := &http.Server{
		Addr:    cfg.Server.Addr(),
		Handler: appHandler,
		BaseContext: func(_ net.Listener) context.Context {
			return gCtx
		},
	}

	g.Go(func() error {
		l.Info("Starting main server", slog.String("addr", cfg.Server.Addr()))
		if err := mainSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	// Observability server
	obsSrv := &http.Server{
		Addr: cfg.Obs.Addr(),
		BaseContext: func(_ net.Listener) context.Context {
			return gCtx
		},
		// TODO: add prometheus metrics handler
	}

	g.Go(func() error {
		l.Info("Starting observability server", slog.String("addr", cfg.Obs.Addr()))
		if err := obsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	// Graceful shutdown
	g.Go(func() error {
		<-gCtx.Done()
		l.Info("Shutting down servers...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		_ = mainSrv.Shutdown(shutdownCtx)
		_ = obsSrv.Shutdown(shutdownCtx)
		return nil
	})

	return g.Wait()
}
