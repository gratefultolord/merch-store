package models

type User struct {
	ID           int                 `db:"id"`
	Username     string              `db:"username"`
	PasswordHash string              `db:"password_hash"`
	Balance      int                 `db:"balance"`
	Inventory    []UserInventoryItem `db:"-"`
}
