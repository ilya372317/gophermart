package main

import (
	"log"
	"net/http"

	"github.com/ilya372317/gophermart/internal/logger"
	"github.com/ilya372317/gophermart/internal/router"
)

const logPath = "./log.txt"

func main() {
	logErr := logger.Initialize(logPath)
	if logErr != nil {
		log.Fatal("failed init logger")
	}
	logger.Log.Info("server is starting...")
	if err := http.ListenAndServe(":8080", router.New()); err != nil {
		logger.Log.Fatalf("failed start server: %v", err)
	}
}
