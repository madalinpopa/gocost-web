package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type SQLiteTrackingRepository struct {
	db DBExecutor
}

func NewSQLiteTrackingRepository(db DBExecutor) *SQLiteTrackingRepository {
	return &SQLiteTrackingRepository{db: db}
}

func (r *SQLiteTrackingRepository) Save(ctx context.Context, group tracking.Group) error {
	groupQuery := `
		INSERT INTO groups (id, user_id, name, description, display_order)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id = excluded.user_id,
			name = excluded.name,
			description = excluded.description,
			display_order = excluded.display_order
	`
	_, err := r.db.ExecContext(ctx, groupQuery,
		group.ID.String(),
		group.UserID.String(),
		group.Name.Value(),
		group.Description.Value(),
		group.Order.Value(),
	)
	if err != nil {
		return err
	}

	categoryQuery := `
		INSERT INTO categories (id, group_id, name, description, is_recurrent, start_month, end_month, budget)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			group_id = excluded.group_id,
			name = excluded.name,
			description = excluded.description,
			is_recurrent = excluded.is_recurrent,
			start_month = excluded.start_month,
			end_month = excluded.end_month,
			budget = excluded.budget,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, category := range group.Categories {
		endMonth := sql.NullString{}
		if !category.EndMonth.IsZero() {
			endMonth = sql.NullString{String: category.EndMonth.Value(), Valid: true}
		}

		_, err = r.db.ExecContext(ctx, categoryQuery,
			category.ID.String(),
			category.GroupID.String(),
			category.Name.Value(),
			category.Description.Value(),
			category.IsRecurrent,
			category.StartMonth.Value(),
			endMonth,
			category.Budget.Cents(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SQLiteTrackingRepository) FindByID(ctx context.Context, id tracking.ID) (tracking.Group, error) {
	groupQuery := `SELECT id, user_id, name, description, display_order FROM groups WHERE id = ?`

	var idStr, userIDStr, nameStr, descriptionStr string
	var orderInt int
	err := r.db.QueryRowContext(ctx, groupQuery, id.String()).Scan(&idStr, &userIDStr, &nameStr, &descriptionStr, &orderInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tracking.Group{}, tracking.ErrGroupNotFound
		}
		return tracking.Group{}, err
	}

	group, err := r.mapToGroup(idStr, userIDStr, nameStr, descriptionStr, orderInt)
	if err != nil {
		return tracking.Group{}, err
	}

	categories, err := r.findCategoriesByGroupID(ctx, group.ID.String())
	if err != nil {
		return tracking.Group{}, err
	}

	for _, category := range categories {
		if err := group.AddCategory(category); err != nil {
			return tracking.Group{}, err
		}
	}

	return *group, nil
}

func (r *SQLiteTrackingRepository) FindByUserID(ctx context.Context, userID tracking.ID) ([]tracking.Group, error) {
	groupQuery := `SELECT id, user_id, name, description, display_order FROM groups WHERE user_id = ? ORDER BY display_order, name`

	rows, err := r.db.QueryContext(ctx, groupQuery, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]*tracking.Group, 0)
	groupByID := make(map[string]*tracking.Group)

	for rows.Next() {
		var idStr, userIDStr, nameStr, descriptionStr string
		var orderInt int
		if err := rows.Scan(&idStr, &userIDStr, &nameStr, &descriptionStr, &orderInt); err != nil {
			return nil, err
		}

		group, err := r.mapToGroup(idStr, userIDStr, nameStr, descriptionStr, orderInt)
		if err != nil {
			return nil, err
		}

		groupByID[group.ID.String()] = group
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []tracking.Group{}, nil
	}

	groupIDs := make([]any, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID.String())
	}

	categoryQuery := buildCategoriesByGroupIDsQuery(len(groupIDs))
	categoryRows, err := r.db.QueryContext(ctx, categoryQuery, groupIDs...)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var (
			idStr, groupIDStr, nameStr, descriptionStr, startMonthStr, currencyStr string
			isRecurrentInt                                                         int
			budgetCents                                                            int64
			endMonth                                                               sql.NullString
		)

		if err := categoryRows.Scan(&idStr, &groupIDStr, &nameStr, &descriptionStr, &isRecurrentInt, &startMonthStr, &endMonth, &budgetCents, &currencyStr); err != nil {
			return nil, err
		}

		category, err := r.mapToCategory(idStr, groupIDStr, nameStr, descriptionStr, isRecurrentInt == 1, startMonthStr, endMonth, budgetCents, currencyStr)
		if err != nil {
			return nil, err
		}

		group, ok := groupByID[groupIDStr]
		if !ok {
			return nil, tracking.ErrGroupNotFound
		}

		if err := group.AddCategory(category); err != nil {
			return nil, err
		}
	}

	if err := categoryRows.Err(); err != nil {
		return nil, err
	}

	result := make([]tracking.Group, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	return result, nil
}

func (r *SQLiteTrackingRepository) FindByUserIDAndMonth(ctx context.Context, userID tracking.ID, month string) ([]tracking.Group, error) {
	groupQuery := `SELECT id, user_id, name, description, display_order FROM groups WHERE user_id = ? ORDER BY display_order, name`

	rows, err := r.db.QueryContext(ctx, groupQuery, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]*tracking.Group, 0)
	groupByID := make(map[string]*tracking.Group)

	for rows.Next() {
		var idStr, userIDStr, nameStr, descriptionStr string
		var orderInt int
		if err := rows.Scan(&idStr, &userIDStr, &nameStr, &descriptionStr, &orderInt); err != nil {
			return nil, err
		}

		group, err := r.mapToGroup(idStr, userIDStr, nameStr, descriptionStr, orderInt)
		if err != nil {
			return nil, err
		}

		groupByID[group.ID.String()] = group
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []tracking.Group{}, nil
	}

	groupIDs := make([]any, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID.String())
	}

	categoryQuery := buildCategoriesByGroupIDsAndMonthQuery(len(groupIDs))
	args := append(groupIDs, month, month, month)
	categoryRows, err := r.db.QueryContext(ctx, categoryQuery, args...)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		var (
			idStr, groupIDStr, nameStr, descriptionStr, startMonthStr, currencyStr string
			isRecurrentInt                                                         int
			budgetCents                                                            int64
			endMonth                                                               sql.NullString
		)

		if err := categoryRows.Scan(&idStr, &groupIDStr, &nameStr, &descriptionStr, &isRecurrentInt, &startMonthStr, &endMonth, &budgetCents, &currencyStr); err != nil {
			return nil, err
		}

		category, err := r.mapToCategory(idStr, groupIDStr, nameStr, descriptionStr, isRecurrentInt == 1, startMonthStr, endMonth, budgetCents, currencyStr)
		if err != nil {
			return nil, err
		}

		group, ok := groupByID[groupIDStr]
		if !ok {
			return nil, tracking.ErrGroupNotFound
		}

		if err := group.AddCategory(category); err != nil {
			return nil, err
		}
	}

	if err := categoryRows.Err(); err != nil {
		return nil, err
	}

	result := make([]tracking.Group, 0, len(groups))
	for _, group := range groups {
		result = append(result, *group)
	}

	return result, nil
}

func (r *SQLiteTrackingRepository) FindGroupByCategoryID(ctx context.Context, categoryID tracking.ID) (tracking.Group, error) {
	query := `
		SELECT g.id, g.user_id, g.name, g.description, g.display_order
		FROM groups g
		JOIN categories c ON g.id = c.group_id
		WHERE c.id = ?
	`
	var idStr, userIDStr, nameStr, descriptionStr string
	var orderInt int
	err := r.db.QueryRowContext(ctx, query, categoryID.String()).Scan(&idStr, &userIDStr, &nameStr, &descriptionStr, &orderInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tracking.Group{}, tracking.ErrGroupNotFound
		}
		return tracking.Group{}, err
	}

	group, err := r.mapToGroup(idStr, userIDStr, nameStr, descriptionStr, orderInt)
	if err != nil {
		return tracking.Group{}, err
	}

	categories, err := r.findCategoriesByGroupID(ctx, group.ID.String())
	if err != nil {
		return tracking.Group{}, err
	}

	for _, category := range categories {
		if err := group.AddCategory(category); err != nil {
			return tracking.Group{}, err
		}
	}

	return *group, nil
}

func (r *SQLiteTrackingRepository) Delete(ctx context.Context, id tracking.ID) error {
	query := `DELETE FROM groups WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return tracking.ErrGroupNotFound
	}
	return nil
}

func (r *SQLiteTrackingRepository) DeleteCategory(ctx context.Context, id tracking.ID) error {
	query := `DELETE FROM categories WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return tracking.ErrCategoryNotFound
	}
	return nil
}

func (r *SQLiteTrackingRepository) findCategoriesByGroupID(ctx context.Context, groupID string) ([]*tracking.Category, error) {
	query := `
		SELECT c.id, c.group_id, c.name, c.description, c.is_recurrent, c.start_month, c.end_month, c.budget, u.currency
		FROM categories c
		JOIN groups g ON c.group_id = g.id
		JOIN users u ON g.user_id = u.id
		WHERE c.group_id = ?
		ORDER BY c.name
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*tracking.Category
	for rows.Next() {
		var (
			idStr, groupIDStr, nameStr, descriptionStr, startMonthStr, currencyStr string
			isRecurrentInt                                                         int
			budgetCents                                                            int64
			endMonth                                                               sql.NullString
		)

		if err := rows.Scan(&idStr, &groupIDStr, &nameStr, &descriptionStr, &isRecurrentInt, &startMonthStr, &endMonth, &budgetCents, &currencyStr); err != nil {
			return nil, err
		}

		category, err := r.mapToCategory(idStr, groupIDStr, nameStr, descriptionStr, isRecurrentInt == 1, startMonthStr, endMonth, budgetCents, currencyStr)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *SQLiteTrackingRepository) mapToGroup(idStr, userIDStr, nameStr, descriptionStr string, orderInt int) (*tracking.Group, error) {
	id, err := identifier.ParseID(idStr)
	if err != nil {
		return nil, err
	}

	userID, err := identifier.ParseID(userIDStr)
	if err != nil {
		return nil, err
	}

	name, err := tracking.NewNameVO(nameStr)
	if err != nil {
		return nil, err
	}

	description, err := tracking.NewDescriptionVO(descriptionStr)
	if err != nil {
		return nil, err
	}

	order, err := tracking.NewOrderVO(orderInt)
	if err != nil {
		return nil, err
	}

	return tracking.NewGroup(id, userID, name, description, order), nil
}

func (r *SQLiteTrackingRepository) mapToCategory(idStr, groupIDStr, nameStr, descriptionStr string, isRecurrent bool, startMonthStr string, endMonth sql.NullString, budgetCents int64, currencyStr string) (*tracking.Category, error) {
	id, err := identifier.ParseID(idStr)
	if err != nil {
		return nil, err
	}

	groupID, err := identifier.ParseID(groupIDStr)
	if err != nil {
		return nil, err
	}

	name, err := tracking.NewNameVO(nameStr)
	if err != nil {
		return nil, err
	}

	description, err := tracking.NewDescriptionVO(descriptionStr)
	if err != nil {
		return nil, err
	}

	startMonth, err := tracking.ParseMonth(startMonthStr)
	if err != nil {
		return nil, err
	}

	var endMonthValue tracking.Month
	if endMonth.Valid {
		endMonthValue, err = tracking.ParseMonth(endMonth.String)
		if err != nil {
			return nil, err
		}
	}

	budget, err := money.New(budgetCents, currencyStr)
	if err != nil {
		return nil, err
	}

	return tracking.NewCategory(id, groupID, name, description, isRecurrent, startMonth, endMonthValue, budget)
}

func buildCategoriesByGroupIDsQuery(count int) string {
	placeholders := strings.Repeat("?,", count)
	placeholders = strings.TrimSuffix(placeholders, ",")
	return fmt.Sprintf(`
		SELECT c.id, c.group_id, c.name, c.description, c.is_recurrent, c.start_month, c.end_month, c.budget, u.currency
		FROM categories c
		JOIN groups g ON c.group_id = g.id
		JOIN users u ON g.user_id = u.id
		WHERE c.group_id IN (%s)
		ORDER BY c.group_id, c.name
	`, placeholders)
}

func buildCategoriesByGroupIDsAndMonthQuery(count int) string {
	placeholders := strings.Repeat("?,", count)
	placeholders = strings.TrimSuffix(placeholders, ",")
	return fmt.Sprintf(`
		SELECT c.id, c.group_id, c.name, c.description, c.is_recurrent, c.start_month, c.end_month, c.budget, u.currency
		FROM categories c
		JOIN groups g ON c.group_id = g.id
		JOIN users u ON g.user_id = u.id
		WHERE c.group_id IN (%s) AND (
			(c.is_recurrent = 1 AND c.start_month <= ? AND (c.end_month IS NULL OR c.end_month = '' OR c.end_month >= ?))
			OR (c.is_recurrent = 0 AND c.start_month = ?)
		)
		ORDER BY c.group_id, c.name
	`, placeholders)
}
