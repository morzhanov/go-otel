package model

type Payment struct {
	ID      string `json:"id,omitempty"`
	OrderID string `json:"orderId,omitempty"`
	Name    string `json:"name,omitempty"`
	Amount  int32  `json:"amount,omitempty"`
	Status  string `json:"status,omitempty"`
}
