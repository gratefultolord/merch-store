package services

import (
	"context"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/gratefultolord/merch-store/internal/repository"
)

type InfoService interface {
	GetUserInfo(ctx context.Context, userID int) (*models.InfoResponse, error)
}

type infoService struct {
	userRepo    repository.UserRepo
	coinService CoinService
}

func NewInfoService(userRepo repository.UserRepo, coinService CoinService) InfoService {
	return &infoService{
		userRepo:    userRepo,
		coinService: coinService,
	}
}

func (s *infoService) GetUserInfo(ctx context.Context, userID int) (*models.InfoResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("service: failed getting user info by id: %v", err)
	}

	if user == nil {
		return nil, fmt.Errorf("service: user not found")
	}

	coinHistory, err := s.coinService.GetCoinHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service: failed getting coin history: %v", err)
	}

	inventory := convertInventory(user.Inventory)
	fmt.Printf("services: inventory value: %+v\n", inventory)

	response := &models.InfoResponse{
		Coins:       user.Balance,
		Inventory:   inventory,
		CoinHistory: *coinHistory,
	}

	fmt.Printf("services: response value: %+v\n", response)

	return response, nil
}

func convertInventory(inventory []models.UserInventoryItem) []models.UserInventoryItemResponse {
	result := make([]models.UserInventoryItemResponse, len(inventory))
	for i, item := range inventory {
		result[i] = models.UserInventoryItemResponse{
			Type:     item.Type.Name,
			Quantity: item.Quantity,
		}
	}
	return result
}
