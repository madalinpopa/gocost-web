package sqlite_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/madalinpopa/gocost-web/migrations"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

// openDB initializes and returns an in-memory SQLite database or an apperror if the connection fails.
func openDB() (db *sql.DB, err error) {
	testDB, err := sql.Open("sqlite3", ":memory:?_foreign_keys=1")
	if err != nil {
		return nil, err
	}
	return testDB, nil
}

// initTestDB initializes and returns an in-memory SQLite test database with applied migrations using Goose.
func initTestDB() (db *sql.DB, err error) {
	testDB, err := openDB()
	if err != nil {
		return nil, err
	}
	goose.SetBaseFS(migrations.MigrationFiles)
	if err = goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}
	if err = goose.Up(testDB, "."); err != nil {
		return nil, err
	}

	return testDB, nil
}

// newTestDB creates and returns a new in-memory SQLite test database, exiting the program if initialization fails.
func newTestDB() *sql.DB {
	testDB, err := initTestDB()
	if err != nil {
		fmt.Println("failed to get database connection", err)
		os.Exit(1)
	}
	return testDB
}

// closeDB safely closes the provided database connection, logging any errors encountered during closure.
func closeDB(db *sql.DB) {
	if db != nil {
		err := db.Close()
		if err != nil {
			fmt.Println("failed to close database connection", err)
			os.Exit(1)
		}
	}
}

func TestMain(m *testing.M) {
	testDB = newTestDB()
	defer closeDB(testDB)
	os.Exit(m.Run())
}

func createRandomUser(t *testing.T) *identity.User {
	t.Helper()
	id, err := identifier.NewID()
	assert.NoError(t, err)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	suffix := r.Int()

	username, err := identity.NewUsernameVO(fmt.Sprintf("user%d", suffix))
	assert.NoError(t, err)

	email, err := identity.NewEmailVO(fmt.Sprintf("user%d@example.com", suffix))
	assert.NoError(t, err)

	password, err := identity.NewPasswordVO("$2a$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12")
	assert.NoError(t, err)

	currency, err := identity.NewCurrencyVO("USD")
	assert.NoError(t, err)

	return identity.NewUser(id, username, email, password, currency)
}

func createRandomIncome(t *testing.T, userID identifier.ID) *income.Income {
	t.Helper()
	id, err := identifier.NewID()
	require.NoError(t, err)

	amount, err := money.New(1000, "USD")
	require.NoError(t, err)

	source, err := income.NewSourceVO("Salary")
	require.NoError(t, err)

	return &income.Income{
		ID:         id,
		UserID:     userID,
		Amount:     amount,
		Source:     source,
		ReceivedAt: time.Now(),
	}
}

func createRandomGroup(t *testing.T, userID identifier.ID) *tracking.Group {
	t.Helper()
	id, err := identifier.NewID()
	require.NoError(t, err)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	suffix := r.Int()

	name, err := tracking.NewNameVO(fmt.Sprintf("Group %d", suffix))
	require.NoError(t, err)

	description, err := tracking.NewDescriptionVO(fmt.Sprintf("Description for group %d", suffix))
	require.NoError(t, err)

	order, err := tracking.NewOrderVO(0)
	require.NoError(t, err)

	group := tracking.NewGroup(id, userID, name, description, order)

	query := `INSERT INTO groups (id, user_id, name, description, display_order) VALUES (?, ?, ?, ?, ?)`
	_, err = testDB.Exec(query, group.ID.String(), group.UserID.String(), group.Name.Value(), group.Description.Value(), group.Order.Value())
	require.NoError(t, err)

	return group
}

func createRandomCategory(t *testing.T, groupID identifier.ID) *tracking.Category {
	t.Helper()
	id, err := identifier.NewID()
	require.NoError(t, err)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	suffix := r.Int()

	name, err := tracking.NewNameVO(fmt.Sprintf("Category %d", suffix))
	require.NoError(t, err)

	description, err := tracking.NewDescriptionVO(fmt.Sprintf("Description for category %d", suffix))
	require.NoError(t, err)

	startMonth := tracking.NewMonthFromTime(time.Now())
	category, err := tracking.NewCategory(id, groupID, name, description, false, startMonth, tracking.Month{}, money.Money{})
	require.NoError(t, err)

	endMonth := sql.NullString{}
	if !category.EndMonth.IsZero() {
		endMonth = sql.NullString{String: category.EndMonth.Value(), Valid: true}
	}

	query := `INSERT INTO categories (id, group_id, name, description, is_recurrent, start_month, end_month) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err = testDB.Exec(
		query,
		category.ID.String(),
		category.GroupID.String(),
		category.Name.Value(),
		category.Description.Value(),
		category.IsRecurrent,
		category.StartMonth.Value(),
		endMonth,
	)
	require.NoError(t, err)

	return category
}

func createRandomExpense(t *testing.T, categoryID identifier.ID) *expense.Expense {
	t.Helper()
	id, err := identifier.NewID()
	require.NoError(t, err)

	amount, err := money.New(500, "USD")
	require.NoError(t, err)

	description, err := expense.NewExpenseDescriptionVO("Lunch")
	require.NoError(t, err)

	payment := expense.NewUnpaidStatus()
	exp, err := expense.NewExpense(id, categoryID, amount, description, time.Now(), payment)
	require.NoError(t, err)

	return exp
}

func newGroupWithID(t *testing.T, id identifier.ID, userID identifier.ID, name string) *tracking.Group {
	t.Helper()
	nameVO, err := tracking.NewNameVO(name)
	require.NoError(t, err)
	descriptionVO, err := tracking.NewDescriptionVO(fmt.Sprintf("%s description", name))
	require.NoError(t, err)
	orderVO, err := tracking.NewOrderVO(0)
	require.NoError(t, err)
	return tracking.NewGroup(id, userID, nameVO, descriptionVO, orderVO)
}

func addCategory(t *testing.T, group *tracking.Group, name string, isRecurrent bool, startMonth tracking.Month, endMonth tracking.Month) *tracking.Category {
	t.Helper()
	id, err := identifier.NewID()
	require.NoError(t, err)
	return addCategoryWithID(t, group, id, name, isRecurrent, startMonth, endMonth)
}

func addCategoryWithID(t *testing.T, group *tracking.Group, id identifier.ID, name string, isRecurrent bool, startMonth tracking.Month, endMonth tracking.Month) *tracking.Category {
	t.Helper()
	nameVO, err := tracking.NewNameVO(name)
	require.NoError(t, err)
	descriptionVO, err := tracking.NewDescriptionVO(fmt.Sprintf("%s description", name))
	require.NoError(t, err)
	category, err := group.CreateCategory(id, nameVO, descriptionVO, isRecurrent, startMonth, endMonth, money.Money{})
	require.NoError(t, err)
	return category
}

func mustMonth(t *testing.T, year int, month time.Month) tracking.Month {
	t.Helper()
	value, err := tracking.NewMonth(year, month)
	require.NoError(t, err)
	return value
}
