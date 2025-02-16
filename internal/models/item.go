package models

import "fmt"

type Item struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Price int    `db:"price"`
}

//type Inventory struct {
//	Type     string `db: "type"`
//	Quantity int    `db:"quantity"`
//}

func (i Item) String() string {
	return fmt.Sprintf("Item{ID: %d, Name: %s, Price: %d}", i.ID, i.Name, i.Price)
}
