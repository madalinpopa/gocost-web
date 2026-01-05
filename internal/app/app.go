package app

import (
	"log/slog"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/config"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/session"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/response"
)

// ApplicationContext holds the application-wide dependencies. It is used as dependency injection container.
type ApplicationContext struct {
	Config   *config.Config
	Logger   *slog.Logger
	Decoder  *form.Decoder
	Session  session.AuthSessionManager
	Template *response.Template
	Response response.Response
}
