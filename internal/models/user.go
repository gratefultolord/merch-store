package models

import (
	"fmt"
)

// User представляет пользователя
type User struct {
	ID           int                 `db:"id"`
	Username     string              `db:"username"`
	PasswordHash string              `db:"password_hash"`
	Balance      int                 `db:"balance"`
	Inventory    []UserInventoryItem `db:"-"`
}

//// UnmarshalJSON для декодирования JSON в инвентарь
//func (u *User) UnmarshalJSON(data []byte) error {
//	type Alias User
//	aux := &struct {
//		Inventory json.RawMessage `json:"inventory"`
//		*Alias
//	}{
//		Alias: (*Alias)(u),
//	}
//	if err := json.Unmarshal(data, &aux); err != nil {
//		return err
//	}
//	// var inventory []Inventory изменено с []Item
//	if aux.Inventory != nil {
//		var inventory []Item
//		if err := json.Unmarshal(aux.Inventory, &inventory); err != nil {
//			return err
//		}
//		u.Inventory = inventory
//	}
//	return nil
//}
//
//// MarshalJSON для кодирования инвентаря в JSON
//func (u *User) MarshalJSON() ([]byte, error) {
//	type Alias User
//	return json.Marshal(&struct {
//		Inventory []Item `json:"inventory"`
//		*Alias
//	}{
//		Inventory: u.Inventory,
//		Alias:     (*Alias)(u),
//	})
//}

// String для удобного вывода информации о пользователе
func (u *User) String() string {
	return fmt.Sprintf("User{ID: %d, Username: %s, Balance: %.2f, Inventory: %v}", u.ID, u.Username, u.Balance, u.Inventory)
}
