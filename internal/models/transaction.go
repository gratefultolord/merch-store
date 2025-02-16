package models

type Transaction struct {
	ID         int `db:"id"`
	SenderID   int `db:"sender_id"`
	ReceiverID int `db:"receiver_id"`
	Amount     int `db:"amount"`
}
