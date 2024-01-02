package main

import (
	"log"
	"net/http"

	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/dbmanager"
	"github.com/ilya372317/gophermart/internal/logger"
	"github.com/ilya372317/gophermart/internal/router"
	"github.com/joho/godotenv"
)

const logPath = "./log.txt"
const envPath = "./.env"
const migrationPath = "db/migrations"

func main() {
	run()
}

func run() {
	logErr := logger.Initialize(logPath)
	if logErr != nil {
		log.Fatal("failed init logger")
	}
	if err := godotenv.Load(envPath); err != nil {
		logger.Log.Warnf("failed load enviroment viarable from .env: %v", err)
	}

	gophermartConfig, err := config.New()
	if err != nil {
		logger.Log.Fatalf("failed create config: %v", err)
		return
	}

	db, err := dbmanager.Open(gophermartConfig.DatabaseDSN)
	if err != nil {
		logger.Log.Fatalf("failed open database connection: %v", err)
		return
	}
	if err := dbmanager.RunMigrations(db.DB, migrationPath); err != nil {
		logger.Log.Fatalf("failed run migrations: %v", err)
		return
	}

	defer func() {
		if err = db.Close(); err != nil {
			logger.Log.Fatalf("failed close db connection: %v", err)
		}
	}()

	logger.Log.Infof("server is starting at host: [%s]...", gophermartConfig.Host)
	if err := http.ListenAndServe(gophermartConfig.Host, router.New()); err != nil {
		logger.Log.Fatalf("failed start server: %v", err)
		return
	}
}
