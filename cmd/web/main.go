package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/middleware"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/response"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/router"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

var (
	address string
	port    int
	dsn     string
	env     string
	version string = "dev"
)

type application struct {
	db     *sql.DB
	logger *slog.Logger
	conf   *config.Config
	router *router.Router
}

func newApplication(db *sql.DB, logger *slog.Logger, conf *config.Config) *application {
	tt := response.NewTemplate(logger, conf)
	ss := web.New(db, conf)
	mm := middleware.New(logger, conf, ss)
	r := response.NewResponse(logger)
	fd := form.NewDecoder()

	fd.RegisterCustomTypeFunc(func(vals []string) (any, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	uow := sqlite.NewUnitOfWork(db)

	appContext := app.HandlerContext{
		Config:   conf,
		Logger:   logger,
		Decoder:  fd,
		Template: tt,
		Response: r,
		Session:  ss,
	}

	useCases := usecase.New(uow, logger)
	handlers := handler.New(appContext, useCases)

	rr := router.New(mm)
	rr.RegisterRoutes(handlers)
	return &application{
		db:     db,
		logger: logger,
		conf:   conf,
		router: rr,
	}
}

func (a *application) handler() http.Handler {
	return a.router.Handlers()
}

func init() {
	flag.StringVar(&address, "addr", "0.0.0.0", "httpserver address to listen to")
	flag.IntVar(&port, "port", 4000, "httpserver port to listen to")
	flag.StringVar(&dsn, "dsn", "data.sqlite", "database connection string")
	flag.StringVar(&env, "env", "development", "application environment")
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	conf := config.New().
		WithServerAddr(address, port).
		WithDatabaseDsn(dsn).
		WithEnvironment(env)

	conf.Version = version

	logger.Info("loading environments")
	err := conf.LoadEnvironments()
	if err != nil {
		logger.Error("failed to load environments configuration", "error", err)
		return fmt.Errorf("failed to load environments configuration: %w", err)
	}

	logger.Info("connect to database", "dsn", conf.Dsn)
	db, err := sqlite.NewDatabaseConnection(context.Background(), conf.Dsn)
	if err != nil {
		logger.Error("failed to get database connection", "err", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	application := newApplication(db, logger, conf)

	httpServer := http.Server{
		Addr:         fmt.Sprintf("%s:%d", conf.Addr, conf.Port),
		Handler:      application.handler(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
	}

	logger.Info("Environment", "mode", conf.GetEnvironment())
	logger.Info("Starting httpserver", "addr", conf.Addr, "port", conf.Port)

	// Start httpserver
	if err := httpServer.ListenAndServe(); err != nil {
		logger.Error(err.Error())
		return fmt.Errorf("failed to start httpserver: %w", err)
	}

	return nil
}
