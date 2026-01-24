package handler

import (
	"log/slog"

	"github.com/go-playground/form/v4"
	"github.com/madalinpopa/gocost-web/internal/config"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type PublicHandlers struct {
	IndexHandler    IndexHandler
	LoginHandler    LoginHandler
	LogoutHandler   LogoutHandler
	RegisterHandler RegisterHandler
}

type PrivateHandlers struct {
	HomeHandler     HomeHandler
	IncomeHandler   IncomeHandler
	GroupHandler    GroupHandler
	CategoryHandler CategoryHandler
	ExpenseHandler  ExpenseHandler
}

type Handlers struct {
	Public  PublicHandlers
	Private PrivateHandlers
}

func New(app HandlerContext, uc *usecase.UseCase) Handlers {
	return Handlers{
		Public: PublicHandlers{
			IndexHandler:    NewIndexHandler(app),
			LoginHandler:    NewLoginHandler(app, uc.AuthUseCase),
			LogoutHandler:   NewLogoutHandler(app, uc.AuthUseCase),
			RegisterHandler: NewRegisterHandler(app, uc.AuthUseCase),
		},
		Private: PrivateHandlers{
			HomeHandler:     NewHomeHandler(app, uc.IncomeUseCase, uc.ExpenseUseCase, uc.GroupUseCase, uc.CategoryUseCase),
			IncomeHandler:   NewIncomeHandler(app, uc.IncomeUseCase, uc.ExpenseUseCase),
			GroupHandler:    NewGroupHandler(app, uc.GroupUseCase),
			CategoryHandler: NewCategoryHandler(app, uc.CategoryUseCase),
			ExpenseHandler:  NewExpenseHandler(app, uc.ExpenseUseCase),
		},
	}
}

// HandlerContext holds the application-wide dependencies. It is used as dependency injection container.
type HandlerContext struct {
	Config   *config.Config
	Logger   *slog.Logger
	Decoder  *form.Decoder
	Session  web.AuthSessionManager
	Template *web.Template
	Errors   respond.ErrorHandler
	Htmx     respond.HtmxHandler
	Notify   respond.NotifyHandler
}
