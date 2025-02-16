package services

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/gratefultolord/merch-store/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Auth(ctx context.Context, username string, password string) (string, error)
}

type authService struct {
	userRepo repository.UserRepo
	secret   string
}

func NewAuthService(userRepo repository.UserRepo, secret string) AuthService {
	return &authService{
		userRepo: userRepo,
		secret:   secret,
	}
}

func (s *authService) Auth(ctx context.Context, username string, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("services: failed to get user by username: %w", err)
	}
	if user == nil {
		newUser := &models.User{
			Username:     username,
			PasswordHash: password,
			Balance:      1000,
		}

		if err := s.userRepo.Create(ctx, newUser); err != nil {
			return "", fmt.Errorf("services: failed to create user: %w", err)
		}

		user = newUser
	} else {
		fmt.Printf("user.PasswordHash: %v, password: %v", user.PasswordHash, password)
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return "", fmt.Errorf("services: invalid password: %w", err)
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
	})

	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("services: failed to sign token: %w", err)
	}

	return tokenString, nil
}
