package apigw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/morzhanov/go-otel/api/grpc/order"
	"github.com/morzhanov/go-otel/api/grpc/payment"
)

type client struct {
	orderUrl      string
	paymentClient payment.PaymentClient
}

type Client interface {
	CreateOrder(msg *order.CreateOrderMessage) (*order.OrderMessage, error)
	ProcessOrder(orderID string) (*order.OrderMessage, error)
	GetPaymentInfo(orderID string) (*payment.PaymentMessage, error)
}

func (c *client) CreateOrder(msg *order.CreateOrderMessage) (*order.OrderMessage, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	res, err := http.Post(c.orderUrl, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	o := order.OrderMessage{}
	if err := json.Unmarshal(body, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (c *client) ProcessOrder(orderID string) (*order.OrderMessage, error) {
	url := fmt.Sprintf("%s/%s", c.orderUrl, orderID)
	res, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	o := order.OrderMessage{}
	if err := json.Unmarshal(body, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (c *client) GetPaymentInfo(orderID string) (*payment.PaymentMessage, error) {
	msg := payment.GetPaymentInfoRequest{OrderId: orderID}
	return c.paymentClient.GetPaymentInfo(context.Background(), &msg)
}

func NewClient(orderUrl string, paymentClient payment.PaymentClient) Client {
	return &client{orderUrl, paymentClient}
}
