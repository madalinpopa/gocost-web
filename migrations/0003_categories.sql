-- +goose Up
CREATE TABLE categories
(
    id           TEXT PRIMARY KEY,
    group_id     TEXT         NOT NULL,
    name         VARCHAR(100) NOT NULL,
    description  TEXT,
    is_recurrent INTEGER      NOT NULL DEFAULT 0,
    start_month  TEXT         NOT NULL DEFAULT (strftime('%Y-%m', 'now')),
    end_month    TEXT,
    budget       INTEGER      NOT NULL DEFAULT 0,
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE
);

CREATE INDEX idx_categories_group_id ON categories(group_id);
CREATE INDEX idx_categories_name ON categories(name);
CREATE INDEX idx_categories_is_recurrent ON categories(is_recurrent);
CREATE INDEX idx_categories_start_month ON categories(start_month);

-- +goose StatementBegin
CREATE TRIGGER trigger_categories_updated_at AFTER UPDATE ON categories FOR EACH ROW
BEGIN
    UPDATE categories SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trigger_categories_updated_at;
DROP INDEX IF EXISTS idx_categories_name;
DROP INDEX IF EXISTS idx_categories_group_id;
DROP INDEX IF EXISTS idx_categories_is_recurrent;
DROP INDEX IF EXISTS idx_categories_start_month;
DROP TABLE IF EXISTS categories;
