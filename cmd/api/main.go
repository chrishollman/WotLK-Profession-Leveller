package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/chrishollman/WotLK-Profession-Leveller/internal/data"
	"github.com/chrishollman/WotLK-Profession-Leveller/internal/tsm"
)

const version = "1.0.0"

type config struct {
	port   int
	env    string
	tsmKey string
}

type application struct {
	config     config
	logger     *zap.SugaredLogger
	stores     *data.Stores
	tsmService *tsm.TSMService
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("unable to start up, couldn't load .env file")
	}

	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|production)")
	flag.Parse()
	cfg.tsmKey = os.Getenv("TSM_API_KEY")

	logger := initLogger(&cfg)

	stores := data.NewStores(logger)

	tsmService := tsm.NewTSMService(cfg.tsmKey, logger)
	tsmService.AuthTicker()

	app := &application{
		config:     cfg,
		logger:     logger,
		stores:     stores,
		tsmService: tsmService,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.logger.Infof("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	app.logger.Fatal(err.Error())
}

// Setup the logging library and configuration based on the provided environment
func initLogger(cfg *config) *zap.SugaredLogger {
	var config zap.Config

	switch cfg.env {
	case "development":
		config = zap.NewDevelopmentConfig()
	default:
		config = zap.NewProductionConfig()
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Errorf("failed to initiate zap logger: %v", err))
	}

	return logger.Sugar()
}
