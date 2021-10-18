package apigw

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzhanov/go-otel/api/grpc/order"
	"github.com/morzhanov/go-otel/internal/metrics"
	"github.com/morzhanov/go-otel/internal/rest"
	"go.uber.org/zap"
)

type controller struct {
	rest.BaseController
	client Client
}

type Controller interface {
	Listen(ctx context.Context, cancel context.CancelFunc, port string)
}

func (c *controller) handleHttpErr(ctx *gin.Context, err error) {
	ctx.String(http.StatusInternalServerError, err.Error())
	c.BaseController.Logger().Info("error in the REST handler", zap.Error(err))
}

func (c *controller) handleCreateOrder(ctx *gin.Context) {
	d := order.CreateOrderMessage{}
	if err := c.BaseController.ParseRestBody(ctx, &d); err != nil {
		c.handleHttpErr(ctx, err)
		return
	}
	res, err := c.client.CreateOrder(&d)
	if err != nil {
		c.handleHttpErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, res)
}

func (c *controller) handleProcessOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	res, err := c.client.ProcessOrder(id)
	if err != nil {
		c.handleHttpErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *controller) handleGetPaymentInfo(ctx *gin.Context) {
	orderID := ctx.Param("orderID")
	res, err := c.client.GetPaymentInfo(orderID)
	if err != nil {
		c.handleHttpErr(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *controller) Listen(
	ctx context.Context,
	cancel context.CancelFunc,
	port string,
) {
	c.BaseController.Listen(ctx, cancel, port)
}

func NewController(
	//tracer opentracing.Tracer,
	client Client,
	log *zap.Logger,
	mc metrics.Collector,
) Controller {
	bc := rest.NewBaseController(log, mc)
	c := controller{BaseController: bc, client: client}

	r := bc.Router()
	r.POST("/order", bc.Handler(c.handleCreateOrder))
	r.PUT("/order/:id", bc.Handler(c.handleProcessOrder))
	r.GET("/payment/:orderID", bc.Handler(c.handleGetPaymentInfo))
	return &c
}
