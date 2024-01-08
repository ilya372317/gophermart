package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/dbmanager"
	"github.com/ilya372317/gophermart/internal/logger"
	"github.com/ilya372317/gophermart/internal/router"
	"github.com/ilya372317/gophermart/internal/storage"
	"github.com/joho/godotenv"
)

const logPath = "./log.txt"
const envPath = "./.env"
const migrationPath = "db/migrations"
const secondsForFinishRequests = 5
const secondsForWaitAccrualStopping = 5

func main() {
	run()
}

func run() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	logErr := logger.Initialize(logPath)
	if logErr != nil {
		log.Println("failed init logger")
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

	accrualCommand, err := runAccrual(gophermartConfig)
	if err != nil {
		logger.Log.Fatalf("failed start accrual: %v", err)
		return
	}

	server := http.Server{
		Addr:    gophermartConfig.Host,
		Handler: router.New(storage.New(db), gophermartConfig),
	}

	go func() {
		logger.Log.Infof("server is starting at host: [%s]...", gophermartConfig.Host)
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(http.ErrServerClosed, err) {
				logger.Log.Info("server successfully shutdown. goodbye :)")
				return
			}
			logger.Log.Fatalf("failed start server: %v", err)
			return
		}
	}()

	<-ctx.Done()

	if err := stopAccrual(accrualCommand); err != nil {
		logger.Log.Warnf("failed stop accrual system: %v", err)
	}
	logger.Log.Info("completing requests and stopping server...")

	shutdownCtx, stop := context.WithTimeout(context.Background(), time.Second*secondsForFinishRequests)
	defer stop()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Fatalf("failed shutdown server: %v", err)
		return
	}
}

func stopAccrual(cmd *exec.Cmd) error {
	logger.Log.Info("stopping accrual system...")
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send termination signal to accrual system: %w", err)
	}
	time.Sleep(secondsForWaitAccrualStopping * time.Second)
	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill accrual system: %w", err)
	}

	return nil
}

func runAccrual(gopherConfig *config.GophermartConfig) (*exec.Cmd, error) {
	var binaryFile string
	switch runtime.GOARCH {
	case "arm64":
		binaryFile = "accrual_darwin_arm64"
	case "amd64":
		binaryFile = "accrual_linux_amd_64"
	default:
		binaryFile = "accrual_windows_amd64"
	}

	command := "cmd/accrual/" + binaryFile
	cmd := exec.Command(command, "-a", gopherConfig.AccrualAddress)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed start accraul service: %w", err)
	}

	return cmd, nil
}
