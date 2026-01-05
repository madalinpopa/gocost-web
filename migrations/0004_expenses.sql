-- +goose Up
CREATE TABLE expenses
(
    id          TEXT PRIMARY KEY,
    category_id TEXT     NOT NULL,
    amount      REAL     NOT NULL,
    description VARCHAR(255),
    spent_at    DATETIME NOT NULL,
    is_paid     INTEGER  NOT NULL DEFAULT 0,
    paid_at     DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
);
CREATE INDEX idx_expenses_amount ON expenses(amount);
CREATE INDEX idx_expenses_category_id ON expenses(category_id);
CREATE INDEX idx_expenses_spent_at ON expenses(spent_at);
CREATE INDEX idx_expenses_is_paid ON expenses(is_paid);

-- +goose StatementBegin
CREATE TRIGGER trigger_expenses_updated_at AFTER UPDATE ON expenses FOR EACH ROW
BEGIN
    UPDATE expenses SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trigger_expenses_updated_at;
DROP INDEX IF EXISTS idx_expenses_amount;
DROP INDEX IF EXISTS idx_expenses_category_id;
DROP INDEX IF EXISTS idx_expenses_spent_at;
DROP INDEX IF EXISTS idx_expenses_is_paid;
DROP TABLE IF EXISTS expenses;
