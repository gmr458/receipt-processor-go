package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *app) serve() error {
	app.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.setupRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := app.setupShutdown(app.server, "main server")

	app.logger.Info("starting server", "addr", app.server.Addr, "env", app.config.env)

	err := app.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", app.server.Addr)

	return nil
}

func (app *app) serveDebug() error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	app.debugServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.debugPort),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownError := app.setupShutdown(app.debugServer, "debug server")

	app.logger.Info("starting debug server", "addr", app.debugServer.Addr, "env", app.config.env)

	err := app.debugServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped debug server", "addr", app.debugServer.Addr)

	return nil
}

func (app *app) setupShutdown(server *http.Server, serverName string) chan error {
	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		sgnl := <-quit
		app.logger.Info(fmt.Sprintf("shutting down %s", serverName), "signal", sgnl.String())

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
			return
		}

		app.logger.Info("completing background tasks", "addr", server.Addr)

		app.wg.Wait()
		shutdownError <- nil
	}()

	return shutdownError
}
