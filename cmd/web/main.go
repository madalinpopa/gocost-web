package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
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

func buildHTTPHandler(db *sql.DB, logger *slog.Logger, conf *config.Config) http.Handler {
	templateRenderer := web.NewTemplate(logger, conf)
	sessionManager := web.NewSession(db, conf)
	errHandler := respond.NewErrorHandler(logger)
	notify := respond.NewNotify(logger)
	htmx := respond.NewHtmx(errHandler)
	middleware := web.NewMiddleware(logger, conf, sessionManager, errHandler)
	formDecoder := form.NewDecoder()

	formDecoder.RegisterCustomTypeFunc(func(vals []string) (any, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	unitOfWork := sqlite.NewUnitOfWork(db)

	handlerContext := handler.HandlerContext{
		Config:   conf,
		Logger:   logger,
		Decoder:  formDecoder,
		Template: templateRenderer,
		Errors:   errHandler,
		Htmx:     htmx,
		Notify:   notify,
		Session:  sessionManager,
	}

	useCases := usecase.New(unitOfWork, logger)
	webHandlers := handler.New(handlerContext, useCases)

	httpRouter := router.New(middleware)
	httpRouter.RegisterRoutes(webHandlers)
	return httpRouter.Handlers()
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
	slog.SetDefault(logger)

	conf, err := loadConfig(logger)
	if err != nil {
		return err
	}

	logger.Info("connect to database", "dsn", conf.Dsn)
	db, err := sqlite.NewDatabaseConnection(context.Background(), conf.Dsn)
	if err != nil {
		logger.Error("failed to get database connection", "err", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}()

	httpHandler := buildHTTPHandler(db, logger, conf)
	server := newHTTPServer(conf, logger, httpHandler)

	logger.Info("Environment", "mode", conf.GetEnvironment())
	logger.Info("Starting httpserver", "addr", conf.Addr, "port", conf.Port)

	if err := runHTTPServer(context.Background(), server, logger); err != nil {
		return err
	}

	logger.Info("httpserver shutdown completed")

	return nil
}

func loadConfig(logger *slog.Logger) (*config.Config, error) {
	conf := config.New().
		WithServerAddr(address, port).
		WithDatabaseDsn(dsn).
		WithEnvironment(env)

	conf.Version = version

	logger.Info("loading environments")
	if err := conf.LoadEnvironments(); err != nil {
		logger.Error("failed to load environments configuration", "error", err)
		return nil, fmt.Errorf("failed to load environments configuration: %w", err)
	}

	return conf, nil
}

func newHTTPServer(conf *config.Config, logger *slog.Logger, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", conf.Addr, conf.Port),
		Handler:      handler,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func runHTTPServer(ctx context.Context, srv *http.Server, logger *slog.Logger) error {
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.ListenAndServe()
	}()

	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			logger.Info("httpserver stopped")
			return nil
		}

		logger.Error("httpserver terminated unexpectedly", "error", err)
		return fmt.Errorf("httpserver terminated unexpectedly: %w", err)
	case <-shutdownCtx.Done():
		logger.Info("shutdown signal received", "signal", shutdownCtx.Err())
	}

	shutdownTimeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownTimeoutCtx); err != nil {
		logger.Error("failed to shutdown httpserver", "error", err)
		return fmt.Errorf("failed to shutdown httpserver: %w", err)
	}

	return nil
}
