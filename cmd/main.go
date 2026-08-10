package main

import (
	"context"
	"errors"
	"fmt"
	"go-platform-template/internal/adapter/httpserver"
	"go-platform-template/internal/adapter/postgresadapter"
	"go-platform-template/internal/core/widget"
	"go-platform-template/internal/platform/buildinfo"
	"go-platform-template/internal/platform/config"
	"go-platform-template/internal/platform/logger"
	"go-platform-template/internal/platform/pg"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func runApp() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Init logger
	log := logger.New(cfg.LoggerConfig)
	slog.SetDefault(log)

	// Init database
	pgPool, err := pg.New(ctx, cfg.DatabaseConfig, log)
	if err != nil {
		return fmt.Errorf("error initializing database: %w", err)
	}
	defer pgPool.Close()

	// Check database is alive
	dbPingCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	err = pgPool.Ping(dbPingCtx)
	if err != nil {
		log.Warn("db is unreachable at boot", "err", err)
	}

	// Widget
	widgetRepo := postgresadapter.NewWidgetRepository(pgPool)
	widgetService := widget.NewService(widgetRepo, log)
	widgetHandler := httpserver.NewWidgetHandler(widgetService)

	// Reliability
	reliabilityHandler := httpserver.NewProbeHandler(pgPool)

	// HTTP
	router := httpserver.NewRouter(reliabilityHandler, widgetHandler)
	handler := router.RegisterRoutes()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPConfig.Port),
		Handler:           handler,
		ReadTimeout:       cfg.HTTPConfig.ReadTimeout,
		WriteTimeout:      cfg.HTTPConfig.WriteTimeout,
		ReadHeaderTimeout: cfg.HTTPConfig.ReadHeaderTimeout,
		IdleTimeout:       cfg.HTTPConfig.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPConfig.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("server stopped cleanly")
	return nil

}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		_, _ = fmt.Fprintf(os.Stdout, "go-platform-template by Feedok. Version: %s, Commit: %s, Date: %s, Runtime ver: %s\n",
			buildinfo.Version,
			buildinfo.Commit,
			buildinfo.Date,
			runtime.Version())
		os.Exit(0)
	}

	if err := runApp(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
