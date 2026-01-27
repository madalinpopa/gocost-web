package views

type ExpenseStatus string

const (
	StatusPaid   ExpenseStatus = "Paid"
	StatusUnpaid ExpenseStatus = "Unpaid"
)

type CategoryType string

const (
	TypeMonthly   CategoryType = "This month only"
	TypeRecurrent CategoryType = "Recurrent"
)

type BudgetStatus string

const (
	BudgetStatusUnder BudgetStatus = "under"
	BudgetStatusEqual BudgetStatus = "equal"
	BudgetStatusOver  BudgetStatus = "over"
)

type ExpenseView struct {
	ID          string
	Amount      float64
	Currency    string
	Description string
	Status      ExpenseStatus
	SpentAt     string
	PaidAt      string
}

type CategoryView struct {
	ID          string
	Name        string
	Type        CategoryType
	Description string
	StartMonth  string
	EndMonth    string
	Budget      float64
	Spent       float64
	Currency    string
	Expenses    []ExpenseView

	// Progress Bar & Budget fields
	PaidSpent        float64
	UnpaidSpent      float64
	PaidPercentage   float64
	UnpaidPercentage float64
	IsNearBudget     bool
	IsOverBudget     bool
	OverBudgetAmount float64
	RemainingBudget  float64
}

type GroupView struct {
	ID          string
	Name        string
	Description string
	Order       int
	Categories  []CategoryView
}

type DashboardView struct {
	CurrentMonth        string
	CurrentMonthParam   string
	PrevMonth           string
	NextMonth           string
	TotalIncome         float64
	TotalExpenses       float64
	TotalBudgeted       float64
	TotalBudgetedStatus BudgetStatus
	Balance             float64
	BalanceAbs          float64
	Currency            string
	Groups              []GroupView
}
