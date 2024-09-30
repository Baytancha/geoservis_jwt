package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: app.setupRouter(),
	}

	shutdownErr := make(chan error)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		s := <-sigChan

		app.logger.Info("shutting down", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr <- server.Shutdown(ctx)
	}()

	app.logger.Info("starting server", "addr", server.Addr)

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErr
	if err != nil {
		return err
	}

	app.logger.Info("graceful exit", "addr", server.Addr)
	return nil
}
