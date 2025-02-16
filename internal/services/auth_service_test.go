package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, userID int) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) GetUsernameByID(ctx context.Context, userID int) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID int, amount int) error {
	return nil
}

func (m *MockUserRepo) UpdateInventory(ctx context.Context, tx *sqlx.Tx, userID int, inventory []models.UserInventoryItem) error {
	return nil
}

func (m *MockUserRepo) CheckInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, existingQuantity *int) error {
	return nil
}

func (m *MockUserRepo) AddOrIncrementItemInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error {
	return nil
}

func (m *MockUserRepo) AddToInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error {
	return nil
}

func (m *MockUserRepo) UpdateInventoryQuantity(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error {
	return nil
}

func TestAuthService_Auth(t *testing.T) {
	secret := "your-secret-key"

	tests := []struct {
		name              string
		username          string
		password          string
		mockSetup         func(m *MockUserRepo)
		expectErrorSubstr string
		expectedSubClaim  int
	}{
		{
			name:     "Existing user with correct password",
			username: "user1",
			password: "password123",
			mockSetup: func(m *MockUserRepo) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{
					ID:           1,
					Username:     "user1",
					PasswordHash: string(hashedPassword),
					Balance:      1000,
				}
				m.On("GetByUsername", mock.Anything, "user1").Return(user, nil)
			},
			expectErrorSubstr: "",
			expectedSubClaim:  1,
		},
		{
			name:     "Existing user with incorrect password",
			username: "user1",
			password: "wrongpassword",
			mockSetup: func(m *MockUserRepo) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{
					ID:           1,
					Username:     "user1",
					PasswordHash: string(hashedPassword),
					Balance:      1000,
				}
				m.On("GetByUsername", mock.Anything, "user1").Return(user, nil)
			},
			expectErrorSubstr: "services: invalid password:",
			expectedSubClaim:  0,
		},
		{
			name:     "New user creation",
			username: "newuser",
			password: "newpassword",
			mockSetup: func(m *MockUserRepo) {
				m.On("GetByUsername", mock.Anything, "newuser").Return(nil, nil).Once()
				m.On("Create", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
					return u.Username == "newuser" && u.PasswordHash == "newpassword" && u.Balance == 1000
				})).Return(nil)
			},
			expectErrorSubstr: "",
			expectedSubClaim:  0,
		},
		{
			name:     "New user creation failed - database error",
			username: "newuser",
			password: "newpassword",
			mockSetup: func(m *MockUserRepo) {
				m.On("GetByUsername", mock.Anything, "newuser").Return(nil, nil).Once()
				m.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectErrorSubstr: "services: failed to create user: database error",
			expectedSubClaim:  0,
		},
		{
			name:     "Database error during GetByUsername",
			username: "user1",
			password: "password123",
			mockSetup: func(m *MockUserRepo) {
				m.On("GetByUsername", mock.Anything, "user1").Return(nil, errors.New("database error"))
			},
			expectErrorSubstr: "services: failed to get user by username: database error",
			expectedSubClaim:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepo)
			tt.mockSetup(mockUserRepo)

			authService := NewAuthService(mockUserRepo, secret)
			tokenString, err := authService.Auth(context.Background(), tt.username, tt.password)

			if tt.expectErrorSubstr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrorSubstr)
				assert.Empty(t, tokenString)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tokenString)

				parsedToken, parseErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(secret), nil
				})
				assert.NoError(t, parseErr)

				if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
					subClaimFloat, ok := claims["sub"].(float64)
					assert.True(t, ok, "claim sub must be a number")
					subClaim := int(subClaimFloat)
					assert.Equal(t, tt.expectedSubClaim, subClaim)
				} else {
					t.Error("Invalid token claims")
				}
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}
