package main

import (
	"fmt"

	"github.com/go-fuego/fuego"
)

func (app *app) serve() error {
	app.server = fuego.NewServer(
		fuego.WithAddr(fmt.Sprintf(":%d", app.config.port)),
		fuego.WithoutLogger(),
	)

	app.setupRoutes()

	app.logger.Info("starting server", "details", map[string]string{
		"addr": app.server.Addr,
		"env":  app.config.env,
	})
	err := app.server.Run()
	if err != nil {
		return err
	}

	return nil
}
