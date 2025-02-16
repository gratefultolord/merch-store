package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/jmoiron/sqlx"
)

type TransactionRepo interface {
	Create(ctx context.Context, tx *sqlx.Tx, transaction *models.Transaction) error
	GetByUserID(ctx context.Context, userID int) ([]models.Transaction, error)
}

type transactionRepo struct {
	db *sqlx.DB
}

func NewTransactionRepo(db *sqlx.DB) TransactionRepo {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(ctx context.Context, tx *sqlx.Tx, transaction *models.Transaction) error {
	query := `INSERT INTO transactions (sender_id, receiver_id, amount) VALUES ($1, $2, $3) RETURNING id`
	err := tx.QueryRowContext(
		ctx, query, transaction.SenderID, transaction.ReceiverID, transaction.Amount).Scan(&transaction.ID)
	if err != nil {
		return fmt.Errorf("repository: cannot create transaction: %w", err)
	}

	return nil
}

func (r *transactionRepo) GetByUserID(ctx context.Context, userID int) ([]models.Transaction, error) {
	var query string
	query = `
			SELECT id, sender_id, receiver_id, amount
			FROM transactions
			WHERE (sender_id = $1 OR receiver_id = $1) AND receiver_id != -1
			`

	var transactions []models.Transaction
	err := r.db.SelectContext(ctx, &transactions, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: cannot get transactions: %w", err)
	}
	return transactions, nil
}
