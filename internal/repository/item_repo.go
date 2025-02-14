package repository

import (
	"context"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/jmoiron/sqlx"
)

type ItemRepo interface {
	GetAll(ctx context.Context) ([]models.Item, error)
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
		return nil, fmt.Errorf("failed to get all items: %w", err)
	}
	return items, nil
}
