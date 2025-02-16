package models

type InfoResponse struct {
	Coins       int                         `json:"coins"`
	Inventory   []UserInventoryItemResponse `json:"inventory"`
	CoinHistory CoinHistory                 `json:"history"`
}

type UserInventoryItem struct {
	Type     Item `json:"type"`
	Quantity int  `json:"quantity"`
}

type UserInventoryItemResponse struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []TransactionSummary `json:"received"`
	Sent     []TransactionSummary `json:"sent"`
}

type TransactionSummary struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}
