package model

type Order struct {
	Id     string `json:"id,omitempty" db:"id"`
	Name   string `json:"name,omitempty" db:"name"`
	Status string `json:"status,omitempty" db:"status"`
}
