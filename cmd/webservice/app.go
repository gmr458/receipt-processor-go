package main

import (
	"net/http"
	"sync"

	"github.com/gmr458/receipt-processor/logger"
	"github.com/gmr458/receipt-processor/service"
)

type app struct {
	config  config
	logger  logger.Logger
	server  *http.Server
	service service.Service
	wg      sync.WaitGroup
}
