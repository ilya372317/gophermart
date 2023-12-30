package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func Initialize(logPath string) error {
	if Log != nil {
		return nil
	}
	environment := os.Getenv("ENV")
	var zapConfig zap.Config

	if environment == "prod" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}
	zapConfig.OutputPaths = []string{"stdout", logPath}
	log, err := zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed build sLogger: %w", err)
	}
	Log = log.Sugar()

	return nil
}
