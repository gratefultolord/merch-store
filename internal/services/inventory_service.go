package services

import (
	"context"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/jmoiron/sqlx"
)

type InventoryService interface {
	Buy(ctx context.Context, userID int, itemName string) error
}

type inventoryService struct {
	userRepo        repository.UserRepo
	itemRepo        repository.ItemRepo
	transactionRepo repository.TransactionRepo
	db              *sqlx.DB
}

func NewInventoryService(
	userRepo repository.UserRepo,
	itemRepo repository.ItemRepo,
	transactionRepo repository.TransactionRepo,
	db *sqlx.DB,
) InventoryService {
	return &inventoryService{
		userRepo:        userRepo,
		itemRepo:        itemRepo,
		transactionRepo: transactionRepo,
		db:              db,
	}
}

func (s *inventoryService) Buy(ctx context.Context, userID int, itemName string) error {
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

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("services: failed to get user by id: %w", err)
	}
	if user == nil {
		return fmt.Errorf("services: user not found")
	}

	items, err := s.itemRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("services: failed to get all items: %w", err)
	}

	var targetItem *models.Item
	for _, i := range items {
		if i.Name == itemName {
			targetItem = &i
			break
		}
	}

	if targetItem == nil {
		return fmt.Errorf("services: item not found")
	}

	if user.Balance < targetItem.Price {
		return fmt.Errorf("services: insufficient balance")
	}

	if err := s.userRepo.UpdateBalance(ctx, tx, userID, -targetItem.Price); err != nil {
		return fmt.Errorf("services: failed to update user balance: %w", err)
	}

	user.Inventory = append(user.Inventory, *targetItem)
	if err := s.userRepo.UpdateInventory(ctx, tx, userID, user.Inventory); err != nil {
		return fmt.Errorf("services: failed to update user inventory: %w", err)
	}

	transaction := &models.Transaction{
		SenderID:   userID,
		ReceiverID: 0,
		Amount:     targetItem.Price,
	}

	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return fmt.Errorf("services: failed to create transaction: %w", err)
	}

	return nil
}
