package models

import "fmt"

type Transaction struct {
	ID         int `db:"id"`
	SenderID   int `db:"sender_id"`
	ReceiverID int `db:"receiver_id"`
	Amount     int `db:"amount"`
}

func (t Transaction) String() string {
	return fmt.Sprintf("Transaction{ID: %d, SenderID: %d, ReceiverID: %d, Amount: %d}")
}
