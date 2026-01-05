package handler

import (
	"github.com/madalinpopa/gocost-web/internal/app"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler/private"
	"github.com/madalinpopa/gocost-web/internal/interfaces/web/handler/public"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type PublicHandlers struct {
	IndexHandler    public.IndexHandler
	LoginHandler    public.LoginHandler
	LogoutHandler   public.LogoutHandler
	RegisterHandler public.RegisterHandler
}

type PrivateHandlers struct {
	HomeHandler     private.HomeHandler
	IncomeHandler   private.IncomeHandler
	GroupHandler    private.GroupHandler
	CategoryHandler private.CategoryHandler
	ExpenseHandler  private.ExpenseHandler
}

type Handlers struct {
	Public  PublicHandlers
	Private PrivateHandlers
}

func New(app app.ApplicationContext, uc *usecase.UseCase) Handlers {
	return Handlers{
		Public: PublicHandlers{
			IndexHandler:    public.NewIndexHandler(app),
			LoginHandler:    public.NewLoginHandler(app, uc.AuthUseCase),
			LogoutHandler:   public.NewLogoutHandler(app, uc.AuthUseCase),
			RegisterHandler: public.NewRegisterHandler(app, uc.AuthUseCase),
		},
		Private: PrivateHandlers{
			HomeHandler:     private.NewHomeHandler(app, uc.IncomeUseCase, uc.ExpenseUseCase, uc.GroupUseCase, uc.CategoryUseCase),
			IncomeHandler:   private.NewIncomeHandler(app, uc.IncomeUseCase, uc.ExpenseUseCase),
			GroupHandler:    private.NewGroupHandler(app, uc.GroupUseCase),
			CategoryHandler: private.NewCategoryHandler(app, uc.CategoryUseCase),
			ExpenseHandler:  private.NewExpenseHandler(app, uc.ExpenseUseCase),
		},
	}
}
