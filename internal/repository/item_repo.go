package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/jmoiron/sqlx"
)

type ItemRepo interface {
	GetAll(ctx context.Context) ([]models.Item, error)
	GetItemByName(ctx context.Context, name string) (*models.Item, error)
}

type itemRepo struct {
	db *sqlx.DB
}

func NewItemRepo(db *sqlx.DB) ItemRepo {
	return &itemRepo{db: db}
}

func (r *itemRepo) GetAll(ctx context.Context) ([]models.Item, error) {
	var items []models.Item
	query := "SELECT id, name, price FROM items"
	err := r.db.SelectContext(ctx, &items, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: failed to get all items: %w", err)
	}
	return items, nil
}

func (r *itemRepo) GetItemByName(ctx context.Context, name string) (*models.Item, error) {
	var item models.Item
	query := `SELECT id, name, price FROM items WHERE name = $1`

	err := r.db.GetContext(ctx, &item, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: failed to get item by name: %w", err)
	}
	return &item, nil
}
