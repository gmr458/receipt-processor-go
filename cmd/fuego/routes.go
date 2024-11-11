package main

import "github.com/go-fuego/fuego"

func (app *app) setupRoutes() {
	receipts := fuego.Group(app.server, "/receipts")

	fuego.Post(receipts, "/process", app.handlerProcessReceipts)
	fuego.Get(receipts, "/{id}/points", app.handlerGetPoints)
}
