package services

import (
	"context"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/jmoiron/sqlx"
)

type CoinService interface {
	Send(ctx context.Context, fromUserID int, toUserID int, amount float64) error
}

type coinService struct {
	userRepo        repository.UserRepo
	transactionRepo repository.TransactionRepo
	db              *sqlx.DB
}

func NewCoinService(userRepo repository.UserRepo, transactionRepo repository.TransactionRepo, db *sqlx.DB) CoinService {
	return &coinService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func (s *coinService) Send(ctx context.Context, fromUserID, toUserID int, amount float64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("services: failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	fromUser, err := s.userRepo.GetByID(ctx, fromUserID)
	if err != nil {
		return fmt.Errorf("services: failed to get fromUser by id: %w", err)
	}
	if fromUser == nil {
		return fmt.Errorf("services: fromUser not found")
	}

	toUser, err := s.userRepo.GetByID(ctx, toUserID)
	if err != nil {
		return fmt.Errorf("services: failed to get toUser by id: %w", err)
	}
	if toUser == nil {
		return fmt.Errorf("services: toUser not found")
	}

	if fromUser.Balance < amount {
		return fmt.Errorf("services: insufficient balance")
	}

	if err := s.userRepo.UpdateBalance(ctx, tx, fromUserID, -amount); err != nil {
		return fmt.Errorf("services: failed to update balance of fromUser: %w", err)
	}

	if err := s.userRepo.UpdateBalance(ctx, tx, toUserID, amount); err != nil {
		return fmt.Errorf("services: failed to update balance of toUser: %w", err)
	}

	transaction := &models.Transaction{
		SenderID:   fromUserID,
		ReceiverID: toUserID,
		Amount:     amount,
	}

	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return fmt.Errorf("services: failed to create transaction: %w", err)
	}

	return nil
}
