package payment

import (
	"github.com/jmoiron/sqlx"
	gpayment "github.com/morzhanov/go-otel/api/grpc/payment"
	uuid "github.com/satori/go.uuid"
)

type pay struct {
	db *sqlx.DB
}

type Payment interface {
	GetPaymentInfo(in *gpayment.GetPaymentInfoRequest) (*gpayment.PaymentMessage, error)
	ProcessPayment(in *gpayment.ProcessPaymentMessage) error
}

func (p *pay) GetPaymentInfo(in *gpayment.GetPaymentInfoRequest) (*gpayment.PaymentMessage, error) {
	var (
		id, orderID, name, status string
		amount                    int32
	)
	if err := p.db.QueryRow(`SELECT * FROM payments WHERE order_id = $id`, in.OrderId).Scan(&id, &orderID, &name, &amount, &status); err != nil {
		return nil, err
	}
	return &gpayment.PaymentMessage{Id: id, OrderId: orderID, Name: name, Status: status, Amount: amount}, nil
}

func (p *pay) ProcessPayment(in *gpayment.ProcessPaymentMessage) error {
	id := uuid.NewV4().String()
	if _, err := p.db.Query(
		`INSERT INTO payments (id, order_id, name, amount, status) VALUES ($id, $orderId, $name, $amount, $status)`,
		id, in.OrderId, in.Name, in.Amount, in.Status,
	); err != nil {
		return err
	}
	return nil
}

func NewPayment(db *sqlx.DB) Payment {
	return &pay{db}
}
