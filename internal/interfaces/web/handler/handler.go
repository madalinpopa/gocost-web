package handler

import (
	"github.com/madalinpopa/gocost-web/internal/app"
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

func New(app app.HandlerContext, uc *usecase.UseCase) Handlers {
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
