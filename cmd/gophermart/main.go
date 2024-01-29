package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"path/filepath"
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

const logPath = "/log.txt"
const envPath = "/.env"
const migrationPath = "/db/migrations"
const secondsForFinishRequests = 5

var (
	_, b, _, _ = runtime.Caller(0)
	root       = filepath.Join(filepath.Dir(b), "../..")
)

func main() {
	run()
}

func run() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	logErr := logger.Initialize(root + logPath)
	if logErr != nil {
		log.Println("failed init logger")
	}
	if err := godotenv.Load(root + envPath); err != nil {
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
	if err := dbmanager.RunMigrations(db.DB, root+migrationPath); err != nil {
		logger.Log.Fatalf("failed run migrations: %v", err)
		return
	}

	defer func() {
		if err = db.Close(); err != nil {
			logger.Log.Fatalf("failed close db connection: %v", err)
		}
	}()

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
