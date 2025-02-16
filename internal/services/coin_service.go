package services

import (
	"context"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/jmoiron/sqlx"
)

type CoinService interface {
	Send(ctx context.Context, fromUserID int, toUserID int, amount int) error
	GetCoinHistory(ctx context.Context, userID int) (*models.CoinHistory, error)
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

func (s *coinService) Send(ctx context.Context, fromUserID, toUserID int, amount int) error {
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

func (s *coinService) GetCoinHistory(ctx context.Context, userID int) (*models.CoinHistory, error) {
	allTransactions, err := s.transactionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("services: failed to get transaction by user id: %w", err)
	}

	received := make([]models.TransactionSummary, 0)
	sent := make([]models.TransactionSummary, 0)

	for _, t := range allTransactions {
		var fromUser, toUser string

		if t.SenderID != -1 {
			fromUser, err = s.userRepo.GetUsernameByID(ctx, t.SenderID)
			if err != nil {
				return nil, fmt.Errorf("services: failed to get sender username by id: %w", err)
			}
		}

		if t.ReceiverID != -1 {
			toUser, err = s.userRepo.GetUsernameByID(ctx, t.ReceiverID)
			if err != nil {
				return nil, fmt.Errorf("services: failed to get receiver username by id: %w", err)
			}
		}

		if t.ReceiverID == userID {
			received = append(received, models.TransactionSummary{
				FromUser: fromUser,
				Amount:   t.Amount,
			})
		} else if t.SenderID == userID {
			sent = append(sent, models.TransactionSummary{
				ToUser: toUser,
				Amount: t.Amount,
			})
		}
	}

	coinHistory := &models.CoinHistory{
		Received: received,
		Sent:     sent,
	}

	return coinHistory, nil
}
