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

	"github.com/ilya372317/gophermart/internal/accrual"
	"github.com/ilya372317/gophermart/internal/config"
	"github.com/ilya372317/gophermart/internal/dbmanager"
	"github.com/ilya372317/gophermart/internal/logger"
	"github.com/ilya372317/gophermart/internal/orderproc"
	"github.com/ilya372317/gophermart/internal/router"
	"github.com/ilya372317/gophermart/internal/storage"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
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

	repo := storage.New(db)
	orderProcessor := orderproc.New(accrual.New(gophermartConfig.AccrualAddress), repo)
	orderProcessor.Start(gophermartConfig)
	go orderProcessor.SupervisingOrders(ctx, gophermartConfig)

	server := &http.Server{
		Addr:    gophermartConfig.Host,
		Handler: router.New(repo, gophermartConfig),
	}

	g := &errgroup.Group{}
	g.Go(func() error {
		return runServer(server, gophermartConfig.Host)
	})
	g.Go(func() error {
		<-ctx.Done()
		return shutdownServer(server)
	})
	g.Go(func() error {
		<-ctx.Done()
		return stopAccrual(accrualCommand)
	})

	if err = g.Wait(); err != nil {
		logger.Log.Warnf("server shutdown with err: %v", err)
	}
}

func shutdownServer(server *http.Server) error {
	shutdownCtx, stop := context.WithTimeout(context.Background(), time.Second*secondsForFinishRequests)
	defer stop()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed shutdown server: %w", err)
	}
	return nil
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
	logger.Log.Info("accrual stopped.")

	return nil
}

func runServer(server *http.Server, host string) error {
	logger.Log.Infof("server is starting at host: [%s]...", host)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(http.ErrServerClosed, err) {
			logger.Log.Info("server successfully shutdown.")
			return nil
		}
		return fmt.Errorf("failed start server: %w", err)
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
