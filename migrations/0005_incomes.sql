-- +goose Up
CREATE TABLE incomes
(
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    amount     REAL NOT NULL,
    source     VARCHAR(255),
    received_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_income_amount ON incomes(amount);
CREATE INDEX idx_income_user_id ON incomes(user_id);
CREATE INDEX idx_income_received_at ON incomes(received_at);

-- +goose StatementBegin
CREATE TRIGGER trigger_incomes_updated_at AFTER UPDATE ON incomes FOR EACH ROW
BEGIN
    UPDATE incomes SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trigger_incomes_updated_at;
DROP INDEX IF EXISTS idx_income_amount;
DROP INDEX IF EXISTS idx_income_user_id;
DROP INDEX IF EXISTS idx_income_received_at;
DROP TABLE IF EXISTS incomes;
