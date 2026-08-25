// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DealUnloker/mockzoo/config"
	"github.com/DealUnloker/mockzoo/internal/controller/restapi"
	persistPetRepo "github.com/DealUnloker/mockzoo/internal/repo/persistent/pet"
	"github.com/DealUnloker/mockzoo/internal/usecase/pet"
	"github.com/DealUnloker/mockzoo/pkg/httpserver"
	"github.com/DealUnloker/mockzoo/pkg/logger"
	"github.com/DealUnloker/mockzoo/pkg/postgres"
)

// Run creates objects via constructors and starts the server.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)
	l.Info("app - Run - starting %s version %s", cfg.App.Name, cfg.App.Version)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	petRepo := persistPetRepo.New(pg)
	petUseCase := pet.New(petRepo)

	// HTTP Server
	httpServer := httpserver.New(l,
		httpserver.Port(cfg.HTTP.Port),
		httpserver.Prefork(cfg.HTTP.UsePreforkMode),
		httpserver.ErrorHandler(restapi.NewErrorHandler()),
	)
	restapi.NewRouter(httpServer.App, petUseCase, l)
	httpServer.Start()

	waitForShutdown(httpServer, l)
}

func waitForShutdown(httpServer *httpserver.Server, l logger.Interface) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	var err error

	select {
	case sig := <-interrupt:
		l.Info("app - Run - signal: %s", sig.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	if err := httpServer.Shutdown(); err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
