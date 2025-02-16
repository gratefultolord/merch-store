package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	GetByID(ctx context.Context, userID int) (*models.User, error)
	GetUsernameByID(ctx context.Context, userID int) (string, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID int, amount int) error
	UpdateInventory(ctx context.Context, tx *sqlx.Tx, userID int, inventory []models.UserInventoryItem) error
	Create(ctx context.Context, user *models.User) error
	CheckInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, existingQuantity *int) error
	AddOrIncrementItemInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error
	AddToInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error
	UpdateInventoryQuantity(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, newQuantity int) error
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) GetByID(ctx context.Context, userID int) (*models.User, error) {
	var user models.User
	query := `
        SELECT u.id, u.username, u.balance,
               i.id AS item_id, i.name AS item_name, ui.quantity AS item_quantity
          FROM users u
          LEFT JOIN user_inventory ui ON u.id = ui.user_id
          LEFT JOIN items i ON ui.item_id = i.id
         WHERE u.id = $1`
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: get user by id failed: %w", err)
	}
	defer rows.Close()

	// Обработка первой строки
	if rows.Next() {
		var itemID sql.NullInt64
		var itemName sql.NullString
		var quantity sql.NullInt64

		err = rows.Scan(&user.ID, &user.Username, &user.Balance, &itemID, &itemName, &quantity)
		if err != nil {
			return nil, fmt.Errorf("repository: scan user row failed: %w", err)
		}
		if itemID.Valid && itemName.Valid && quantity.Valid {
			user.Inventory = append(user.Inventory, models.UserInventoryItem{
				Type: models.Item{
					ID:   int(itemID.Int64),
					Name: itemName.String,
				},
				Quantity: int(quantity.Int64),
			})
		}
	} else {
		// Если данных о пользователе нет
		return nil, nil
	}

	// Обработка последующих строк
	for rows.Next() {
		// Пользовательские данные уже получены, поэтому используем dummy-переменные
		var dummyID int
		var dummyUsername string
		var dummyBalance int
		var itemID sql.NullInt64
		var itemName sql.NullString
		var quantity sql.NullInt64

		err = rows.Scan(&dummyID, &dummyUsername, &dummyBalance, &itemID, &itemName, &quantity)
		if err != nil {
			return nil, fmt.Errorf("repository: scan inventory row failed: %w", err)
		}
		if itemID.Valid && itemName.Valid && quantity.Valid {
			user.Inventory = append(user.Inventory, models.UserInventoryItem{
				Type: models.Item{
					ID:   int(itemID.Int64),
					Name: itemName.String,
				},
				Quantity: int(quantity.Int64),
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: iterate rows error: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetUsernameByID(ctx context.Context, userID int) (string, error) {
	var username string
	query := `
		SELECT username
		FROM users
		WHERE id = $1
		`

	err := r.db.GetContext(ctx, &username, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("repository: get username by id failed: %w", err)
	}
	return username, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, password_hash, balance FROM users WHERE username = $1`
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("repository: get user by username failed: %w", err)
	}
	return &user, nil
}

func (r *userRepo) UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID int, amount int) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return fmt.Errorf("repository: update user balance failed: %w", err)
	}
	return nil
}

func (r *userRepo) UpdateInventory(ctx context.Context, tx *sqlx.Tx, userID int, inventory []models.UserInventoryItem) error {
	for _, item := range inventory {
		var existingQuantity int
		err := r.CheckInventory(ctx, tx, userID, item.Type.ID, &existingQuantity)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("repository: UpdateInventory check failed: %w", err)
		}

		if err == sql.ErrNoRows {
			if err := r.AddToInventory(ctx, tx, userID, item.Type.ID, item.Quantity); err != nil {
				return fmt.Errorf("repository: UpdateInventory add failed: %w", err)
			}
		} else {
			newQuantity := existingQuantity + item.Quantity
			if err := r.UpdateInventoryQuantity(ctx, tx, userID, item.Type.ID, newQuantity); err != nil {
				return fmt.Errorf("repository: UpdateInventory update quantity failed: %w", err)
			}
		}
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
	fmt.Printf("repository: user.ID: %v\n", user.ID)
	if err != nil {
		return fmt.Errorf("repository: failed to create new user: %w", err)
	}
	return nil
}

func (r *userRepo) CheckInventory(
	ctx context.Context, tx *sqlx.Tx,
	userID int, itemID int, existingQuantity *int) error {
	query := `SELECT quantity FROM user_inventory WHERE user_id = $1 AND item_id = $2`
	err := tx.GetContext(ctx, existingQuantity, query, userID, itemID)
	if err != nil {
		return fmt.Errorf("repository: inventory check failed: %w", err)
	}
	return nil
}

func (r *userRepo) AddOrIncrementItemInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error {
	var existingQuantity int
	err := tx.GetContext(ctx, &existingQuantity,
		`SELECT quantity FROM user_inventory WHERE user_id = $1 AND item_id = $2`, userID, itemID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("repository: AddOrIncrement inventory check failed: %w", err)
	}

	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO user_inventory (user_id, item_id, quantity) VALUES ($1, $2, $3)`,
			userID, itemID, quantity)

		if err != nil {
			return fmt.Errorf("repository: AddOrIncrement inventory failed to add: %w", err)
		}
	} else {
		newQuantity := existingQuantity + quantity
		_, err = tx.ExecContext(ctx, `UPDATE user_inventory SET quantity = $1 WHERE user_id = $2 AND item_id = $3`,
			newQuantity, userID, itemID)

		if err != nil {
			return fmt.Errorf("repository: AddOrIncrement inventory failed to update quantity: %w", err)
		}
	}
	return nil
}

func (r *userRepo) AddToInventory(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error {
	query := `INSERT INTO user_inventory (user_id, item_id, quantity) VALUES ($1, $2, $3)`
	_, err := tx.ExecContext(ctx, query, userID, itemID, quantity)
	if err != nil {
		return fmt.Errorf("repository: failed to add item to inventory: %w", err)
	}
	return nil
}

func (r *userRepo) UpdateInventoryQuantity(ctx context.Context, tx *sqlx.Tx, userID int, itemID int, quantity int) error {
	query := `UPDATE user_inventory SET quantity = $1 WHERE user_id = $2 AND item_id = $3`
	_, err := tx.ExecContext(ctx, query, quantity, userID, itemID)
	if err != nil {
		return fmt.Errorf("repository: failed to update item quantity in inventory: %w", err)
	}
	return nil
}
