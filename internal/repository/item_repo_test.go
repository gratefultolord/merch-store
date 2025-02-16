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

func TestItemRepo_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name        string
		mockSetup   func(mock sqlmock.Sqlmock)
		expected    []models.Item
		expectedErr error
	}{
		{
			name: "Items found",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "price"}).
					AddRow(1, "sword", 100).
					AddRow(2, "shield", 200)
				mock.ExpectQuery(`SELECT id, name, price FROM items`).
					WillReturnRows(rows)
			},
			expected: []models.Item{
				{ID: 1, Name: "sword", Price: 100},
				{ID: 2, Name: "shield", Price: 200},
			},
			expectedErr: nil,
		},
		{
			name: "No items found",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "price"})
				mock.ExpectQuery(`SELECT id, name, price FROM items`).
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name: "Database error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, price FROM items`).
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: fmt.Errorf("repository: failed to get all items: %w", sql.ErrConnDone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			repo := NewItemRepo(sqlxDB)
			items, err := repo.GetAll(context.Background())

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, items)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestItemRepo_GetItemByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	tests := []struct {
		name        string
		itemName    string
		mockSetup   func(mock sqlmock.Sqlmock)
		expected    *models.Item
		expectedErr error
	}{
		{
			name:     "Item found",
			itemName: "sword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "price"}).
					AddRow(1, "sword", 100)
				mock.ExpectQuery(`SELECT id, name, price FROM items WHERE name = \$1`).
					WithArgs("sword").
					WillReturnRows(rows)
			},
			expected:    &models.Item{ID: 1, Name: "sword", Price: 100},
			expectedErr: nil,
		},
		{
			name:     "Item not found",
			itemName: "nonexistent_item",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "price"})
				mock.ExpectQuery(`SELECT id, name, price FROM items WHERE name = \$1`).
					WithArgs("nonexistent_item").
					WillReturnRows(rows)
			},
			expected:    nil,
			expectedErr: nil,
		},
		{
			name:     "Database error",
			itemName: "sword",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, price FROM items WHERE name = \$1`).
					WithArgs("sword").
					WillReturnError(sql.ErrConnDone)
			},
			expected:    nil,
			expectedErr: fmt.Errorf("repository: failed to get item by name: %w", sql.ErrConnDone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mock)

			repo := NewItemRepo(sqlxDB)
			item, err := repo.GetItemByName(context.Background(), tt.itemName)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, item)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
