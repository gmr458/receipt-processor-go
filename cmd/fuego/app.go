package main

import (
	"sync"

	"github.com/go-fuego/fuego"

	"github.com/gmr458/receipt-processor/logger"
	"github.com/gmr458/receipt-processor/service"
)

type app struct {
	config      config
	logger      logger.Logger
	server      *fuego.Server
	debugServer *fuego.Server
	service     service.Service
	wg          sync.WaitGroup
}
