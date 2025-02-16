package services

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/gratefultolord/merch-store/mocks"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, sqlMock, err := sqlmock.New()
	assert.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	return sqlxDB, sqlMock
}

func TestCoinService_Send(t *testing.T) {
	sqlxDB, sqlMock := setupTestDB(t)
	defer sqlxDB.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	mockUserRepo := new(mocks.UserRepo)
	mockTransactionRepo := new(mocks.TransactionRepo)

	coinService := NewCoinService(mockUserRepo, mockTransactionRepo, sqlxDB)
	ctx := context.Background()

	fromUser := &models.User{ID: 1, Balance: 100}
	toUser := &models.User{ID: 2, Balance: 50}
	amount := 30

	mockUserRepo.On("GetByID", ctx, 1).Return(fromUser, nil).Once()
	mockUserRepo.On("GetByID", ctx, 2).Return(toUser, nil).Once()
	mockUserRepo.On("UpdateBalance", ctx, mock.Anything, 1, -amount).Return(nil).Once()
	mockUserRepo.On("UpdateBalance", ctx, mock.Anything, 2, amount).Return(nil).Once()
	mockTransactionRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(tr *models.Transaction) bool {
		return tr.SenderID == 1 && tr.ReceiverID == 2 && tr.Amount == amount
	})).Return(nil).Once()

	err := coinService.Send(ctx, 1, 2, amount)
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCoinService_Send_InsufficientBalance(t *testing.T) {
	sqlxDB, sqlMock := setupTestDB(t)
	defer sqlxDB.Close()

	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	mockUserRepo := new(mocks.UserRepo)
	mockTransactionRepo := new(mocks.TransactionRepo)

	coinService := NewCoinService(mockUserRepo, mockTransactionRepo, sqlxDB)
	ctx := context.Background()

	fromUser := &models.User{ID: 1, Balance: 10}
	toUser := &models.User{ID: 2, Balance: 50}
	amount := 30

	mockUserRepo.On("GetByID", ctx, 1).Return(fromUser, nil).Once()
	mockUserRepo.On("GetByID", ctx, 2).Return(toUser, nil).Once()

	err := coinService.Send(ctx, 1, 2, amount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")

	mockUserRepo.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestCoinService_GetCoinHistory(t *testing.T) {
	sqlxDB, _ := setupTestDB(t)
	defer sqlxDB.Close()

	mockUserRepo := new(mocks.UserRepo)
	mockTransactionRepo := new(mocks.TransactionRepo)

	coinService := NewCoinService(mockUserRepo, mockTransactionRepo, sqlxDB)
	ctx := context.Background()

	transactions := []models.Transaction{
		{SenderID: 1, ReceiverID: 2, Amount: 50},
		{SenderID: 2, ReceiverID: 1, Amount: 30},
	}

	mockTransactionRepo.On("GetByUserID", ctx, 1).Return(transactions, nil).Once()
	mockUserRepo.On("GetUsernameByID", ctx, 1).Return("Alice", nil).Twice()
	mockUserRepo.On("GetUsernameByID", ctx, 2).Return("Bob", nil).Twice()

	history, err := coinService.GetCoinHistory(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.Len(t, history.Sent, 1)
	assert.Len(t, history.Received, 1)

	mockTransactionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
