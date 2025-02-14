package repository

import (
	"context"
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
	query := `INSERT INTO transactions (sender_id, receiver_id, amount, timestamp) VALUES ($1, $2, $3, $4)`
	err := tx.QueryRowContext(ctx, query, transaction.SenderID, transaction.ReceiverID, transaction.Amount, transaction.Timestamp).Scan(&transaction.ID)
	if err != nil {
		return fmt.Errorf("cannot create transaction: %w", err)
	}
	return nil
}

func (r *transactionRepo) GetByUserID(ctx context.Context, userID int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := `SELECT id, sender_id, receiver_id, amount, timestamp FROM transactions WHERE sender_id = $1 OR receiver_id = $1`
	err := r.db.SelectContext(ctx, &transactions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("cannot get transactions: %w", err)
	}
	return transactions, nil
}
