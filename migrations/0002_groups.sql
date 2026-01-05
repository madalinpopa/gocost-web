-- +goose Up
CREATE TABLE groups
(
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_groups_name ON groups(name);
CREATE INDEX idx_groups_user_id ON groups(user_id);
CREATE INDEX idx_groups_display_order ON groups(display_order);

-- +goose StatementBegin
CREATE TRIGGER trigger_groups_updated_at AFTER UPDATE ON groups FOR EACH ROW
BEGIN
    UPDATE groups SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS idx_groups_name;
DROP INDEX IF EXISTS idx_groups_user_id;
DROP INDEX IF EXISTS idx_groups_display_order;
DROP TRIGGER IF EXISTS trigger_groups_updated_at;
DROP TABLE IF EXISTS groups;
