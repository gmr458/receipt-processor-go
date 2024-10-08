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

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		sgnl := <-quit
		app.logger.Info("shutting down server", "details", map[string]string{
			"signal": sgnl.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := app.server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", "details", map[string]string{
			"addr": app.server.Addr,
		})

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "details", map[string]string{
		"addr": app.server.Addr,
		"env":  app.config.env,
	})

	err := app.server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "details", map[string]string{
		"addr": app.server.Addr,
	})

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

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		sgnl := <-quit
		app.logger.Info("shutting down debug server", "details", map[string]string{
			"signal": sgnl.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		err := app.debugServer.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("completing background tasks", "details", map[string]string{
			"addr": app.debugServer.Addr,
		})

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting debug server", "details", map[string]string{
		"addr": app.debugServer.Addr,
		"env":  app.config.env,
	})

	err := app.debugServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped debug server", "details", map[string]string{
		"addr": app.debugServer.Addr,
	})

	return nil
}
