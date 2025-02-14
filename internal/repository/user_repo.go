package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, userID int) (*models.User, error)
	UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID int, amount float64) error
	UpdateInventory(ctx context.Context, tx *sqlx.Tx, userID int, inventory []models.Item) error
	Create(ctx context.Context, user *models.User) error
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, balance, inventory FROM users WHERE username = $1`
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: get user by username failed: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByID(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, balance, inventory FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: get user by id failed: %w", err)
	}
	return &user, nil
}

func (r *userRepo) UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID int, amount float64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("repository: update user balance failed: %w", err)
	}
	return nil
}

func (r *userRepo) UpdateInventory(ctx context.Context, tx *sqlx.Tx, userID int, inventory []models.Item) error {
	inventoryJSON, err := json.Marshal(inventory)
	if err != nil {
		return fmt.Errorf("repository: marshal inventory failed: %w", err)
	}

	query := `UPDATE users SET inventory = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, query, string(inventoryJSON), userID)
	if err != nil {
		return fmt.Errorf("repository: update user inventory failed: %w", err)
	}
	return nil
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("repository: hashing password failed: %w", err)
	}

	query := `INSERT INTO users (username, password_hash, balance) VALUES ($1, $2, $3) returning id`
	err = r.db.QueryRowContext(ctx, query, user.Username, hashedPassword, user.Balance).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("repository: failed to create new user: %w", err)
	}
	return nil
}
