package model

type Order struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Amount int32  `json:"amount,omitempty"`
	Status string `json:"status,omitempty"`
}
