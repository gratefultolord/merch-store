package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestUserRepo_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name        string
		userID      int
		mockSetup   func(mock sqlmock.Sqlmock)
		expected    *models.User
		expectedErr error
	}{
		{
			name:   "User found with inventory",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "balance", "item_id", "item_name", "item_quantity"}).
					AddRow(1, "user1", 850, 3, "book", 1).
					AddRow(1, "user1", 850, 4, "sword", 2)
				mock.ExpectQuery(`
									SELECT u\.id, u\.username, u\.balance, i\.id AS item_id, i\.name AS item_name, ui\.quantity AS item_quantity 
									FROM users u LEFT JOIN user_inventory ui ON u\.id = ui\.user_id 
									    LEFT JOIN items i ON ui\.item_id = i\.id WHERE u\.id = \$1
									    `).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expected: &models.User{
				ID:       1,
				Username: "user1",
				Balance:  850,
				Inventory: []models.UserInventoryItem{
					{
						Type: models.Item{
							ID:   3,
							Name: "book",
						},
						Quantity: 1,
					},
					{
						Type: models.Item{
							ID:   4,
							Name: "sword",
						},
						Quantity: 2,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "User found without inventory",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "balance", "item_id", "item_name", "item_quantity"}).
					AddRow(1, "user1", 850, nil, nil, nil)
				mock.ExpectQuery(`
						SELECT u\.id, u\.username, u\.balance, i\.id AS item_id, i\.name AS item_name, ui\.quantity AS item_quantity 
						FROM users u LEFT JOIN user_inventory ui ON u\.id = ui\.user_id 
						    LEFT JOIN items i ON ui\.item_id = i\.id 
						WHERE u\.id = \$1
						`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expected: &models.User{
				ID:        1,
				Username:  "user1",
				Balance:   850,
				Inventory: []models.UserInventoryItem{},
			},
			expectedErr: nil,
		},
		{
			name:   "User not found",
			userID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "balance", "item_id", "item_name", "item_quantity"})
				mock.ExpectQuery(`
						SELECT u\.id, u\.username, u\.balance, i\.id AS item_id, i\.name AS item_name, ui\.quantity AS item_quantity 
						FROM users u LEFT JOIN user_inventory ui ON u\.id = ui\.user_id 
						    LEFT JOIN items i ON ui\.item_id = i\.id 
						WHERE u\.id = \$1
						`).
					WithArgs(999).
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:   "Database error during GetByID",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`
						SELECT u\.id, u\.username, u\.balance, i\.id AS item_id, i\.name AS item_name, ui\.quantity AS item_quantity 
						FROM users u LEFT JOIN user_inventory ui ON u\.id = ui\.user_id 
						    LEFT JOIN items i ON ui\.item_id = i\.id 
						WHERE u\.id = \$1
						`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: fmt.Errorf("repository: get user by id failed: %w", sql.ErrConnDone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			repo := NewUserRepo(sqlxDB)
			user, err := repo.GetByID(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if user != nil && user.Inventory == nil {
				user.Inventory = []models.UserInventoryItem{}
			}
			assert.Equal(t, tt.expected, user)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepo_GetUsernameByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name        string
		userID      int
		mockSetup   func(mock sqlmock.Sqlmock)
		expected    string
		expectedErr error
	}{
		{
			name:   "Username found",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username"}).
					AddRow("user1")
				mock.ExpectQuery(`SELECT username FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expected:    "user1",
			expectedErr: nil,
		},
		{
			name:   "Username not found",
			userID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username"})
				mock.ExpectQuery(`SELECT username FROM users WHERE id = \$1`).
					WithArgs(999).
					WillReturnRows(rows)
			},
			expected:    "",
			expectedErr: nil,
		},
		{
			name:   "Database error",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT username FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    "",
			expectedErr: fmt.Errorf("repository: get username by id failed: %w", sql.ErrConnDone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			repo := NewUserRepo(sqlxDB)
			username, err := repo.GetUsernameByID(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, username)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepo_GetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name        string
		username    string
		mockSetup   func(mock sqlmock.Sqlmock)
		expected    *models.User
		expectedErr error
	}{
		{
			name:     "User found",
			username: "user1",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "balance"}).
					AddRow(1, "user1", "$2a$10$Wn5VZPmD9YRYF4K6T2yv.O3HJ3G2F4T1JG2F4T1JG2F4T1", 1000)
				mock.ExpectQuery(`SELECT id, username, password_hash, balance FROM users WHERE username = \$1`).
					WithArgs("user1").
					WillReturnRows(rows)
			},
			expected: &models.User{
				ID:           1,
				Username:     "user1",
				PasswordHash: "$2a$10$Wn5VZPmD9YRYF4K6T2yv.O3HJ3G2F4T1JG2F4T1JG2F4T1",
				Balance:      1000,
				Inventory:    []models.UserInventoryItem{},
			},
			expectedErr: nil,
		},
		{
			name:     "User not found",
			username: "nonexistent_user",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password_hash", "balance"})
				mock.ExpectQuery(`SELECT id, username, password_hash, balance FROM users WHERE username = \$1`).
					WithArgs("nonexistent_user").
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:     "Database error",
			username: "user1",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, username, password_hash, balance FROM users WHERE username = \$1`).
					WithArgs("user1").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: fmt.Errorf("repository: get user by username failed: %w", sql.ErrConnDone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			repo := NewUserRepo(sqlxDB)
			user, err := repo.GetByUsername(context.Background(), tt.username)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if user != nil && user.Inventory == nil {
				user.Inventory = []models.UserInventoryItem{}
			}
			assert.Equal(t, tt.expected, user)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepo_UpdateBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name        string
		userID      int
		amount      int
		mockSetup   func(mock sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name:   "Successful balance update",
			userID: 1,
			amount: -100,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE users SET balance = balance \+ \$1 WHERE id = \$2`).
					WithArgs(-100, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name:   "Database error during UpdateBalance",
			userID: 1,
			amount: -100,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE users SET balance = balance \+ \$1 WHERE id = \$2`).
					WithArgs(-100, 1).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedErr: fmt.Errorf("repository: update user balance failed: %w", sql.ErrConnDone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			tx, err := sqlxDB.Beginx()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when beginning a transaction", err)
			}

			repo := NewUserRepo(sqlxDB)
			err = repo.UpdateBalance(context.Background(), tx, tt.userID, tt.amount)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
				_ = tx.Rollback()
			} else {
				assert.NoError(t, err)
				if err := tx.Commit(); err != nil {
					t.Fatalf("an error '%s' was not expected when committing a transaction", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
