package app

import (
	"log/slog"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/response"
)

// HandlerContext holds the application-wide dependencies. It is used as dependency injection container.
type HandlerContext struct {
	Config   *config.Config
	Logger   *slog.Logger
	Decoder  *form.Decoder
	Session  web.AuthSessionManager
	Template *response.Template
	Response response.Response
}
