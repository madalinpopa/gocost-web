-- +goose Up
-- Refactor expenses table
CREATE TABLE expenses_new
(
    id          TEXT PRIMARY KEY,
    category_id TEXT     NOT NULL,
    amount      INTEGER  NOT NULL,
    description VARCHAR(255),
    spent_at    DATETIME NOT NULL,
    is_paid     INTEGER  NOT NULL DEFAULT 0,
    paid_at     DATETIME,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE
);

INSERT INTO expenses_new (id, category_id, amount, description, spent_at, is_paid, paid_at, created_at, updated_at)
SELECT id, category_id, CAST(ROUND(amount * 100) AS INTEGER), description, spent_at, is_paid, paid_at, created_at, updated_at
FROM expenses;

DROP TABLE expenses;
ALTER TABLE expenses_new RENAME TO expenses;

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

-- Refactor incomes table
CREATE TABLE incomes_new
(
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    amount     INTEGER NOT NULL,
    source     VARCHAR(255),
    received_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO incomes_new (id, user_id, amount, source, received_at, created_at, updated_at)
SELECT id, user_id, CAST(ROUND(amount * 100) AS INTEGER), source, received_at, created_at, updated_at
FROM incomes;

DROP TABLE incomes;
ALTER TABLE incomes_new RENAME TO incomes;

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
-- Revert incomes table
CREATE TABLE incomes_old
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

INSERT INTO incomes_old (id, user_id, amount, source, received_at, created_at, updated_at)
SELECT id, user_id, CAST(amount AS REAL) / 100.0, source, received_at, created_at, updated_at
FROM incomes;

DROP TABLE incomes;
ALTER TABLE incomes_old RENAME TO incomes;

CREATE INDEX idx_income_amount ON incomes(amount);
CREATE INDEX idx_income_user_id ON incomes(user_id);
CREATE INDEX idx_income_received_at ON incomes(received_at);

-- +goose StatementBegin
CREATE TRIGGER trigger_incomes_updated_at AFTER UPDATE ON incomes FOR EACH ROW
BEGIN
    UPDATE incomes SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- Revert expenses table
CREATE TABLE expenses_old
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

INSERT INTO expenses_old (id, category_id, amount, description, spent_at, is_paid, paid_at, created_at, updated_at)
SELECT id, category_id, CAST(amount AS REAL) / 100.0, description, spent_at, is_paid, paid_at, created_at, updated_at
FROM expenses;

DROP TABLE expenses;
ALTER TABLE expenses_old RENAME TO expenses;

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
