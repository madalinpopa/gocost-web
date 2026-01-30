# Migration to go-money and User-Specific Currency

## 1. Database Migrations (`migrations/`)
- [x] Create `0007_add_currency_to_users.sql`: Add `currency` column to `users` (TEXT, NOT NULL, DEFAULT 'USD').
- [x] Create `0008_refactor_money_columns.sql`:
    - [x] **Expenses:** Convert `amount` from `REAL` to `INTEGER` (cents).
    - [x] **Incomes:** Convert `amount` from `REAL` to `INTEGER` (cents).
    - [x] **Categories:** Ensure `budget` remains consistent (already INTEGER).

## 2. Money Package Refactor (`internal/platform/money/`)
- [ ] Import `github.com/Rhymond/go-money`.
- [ ] Update `Money` struct to wrap `go-money.Money`.
- [ ] Update `New(amount int64, currency string)` and `NewFromFloat(amount float64, currency string)` constructors.
- [ ] Expose/Delegate methods: `Add`, `Subtract`, `Multiply`, `Display`, `Amount` (float representation), `Cents`.

## 3. Repository Layer Updates (`internal/infrastructure/storage/sqlite/`)
- [ ] **Expenses (`SQLiteExpenseRepository`):**
    - [ ] Update queries to fetch `users.currency`.
    - [ ] Update `mapToExpense` to use fetched currency.
    - [ ] Update `Save` to store integer cents.
- [ ] **Incomes (`SQLiteIncomeRepository`):**
    - [ ] Update queries to fetch `users.currency`.
    - [ ] Update mapping and saving logic.
- [ ] **Tracking (`SQLiteTrackingRepository`):**
    - [ ] Join `users` table to fetch currency for Categories (`Budget`).

## 4. Domain & Use Case Updates
- [ ] **Domain Entities:** Ensure factories (`NewExpense`, etc.) accept Currency or Money objects properly constructed.
- [ ] **Use Cases:**
    - [ ] Fetch User (or User's currency) when creating entities.
    - [ ] Pass currency context to `Money` constructors.
    - [ ] Update `Register` logic to use default currency from Config.

## 5. Presentation Layer Updates (`internal/interfaces/web/views/`)
- [ ] **DashboardPresenter:**
    - [ ] Replace `float64` math with `Money` arithmetic.
    - [ ] Use `Money.Display()` for formatting.

## 6. Config Updates (`internal/config/`)
- [ ] Treat `CURRENCY` env var as the system *default* for new users.

## 7. Testing
- [ ] Update `money_test.go` with `go-money` integration.
- [ ] Fix broken tests in other packages due to signature changes.
- [ ] Verify build and test suite pass.
