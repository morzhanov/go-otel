package order

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/morzhanov/go-otel/internal/mq"

	"github.com/gin-gonic/gin"
	porder "github.com/morzhanov/go-otel/api/grpc/order"
	"github.com/morzhanov/go-otel/internal/metrics"
	"github.com/morzhanov/go-otel/internal/rest"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type service struct {
	rest.BaseController
	coll *mongo.Collection
	mq   mq.MQ
}

type Service interface {
	Listen() error
}

func (s *service) handleHttpErr(ctx *gin.Context, err error) {
	ctx.String(http.StatusInternalServerError, err.Error())
	s.BaseController.Logger().Info("error in the REST handler", zap.Error(err))
}

func (s *service) handleCreateOrder(c *gin.Context) {
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return
	}
	d := porder.CreateOrderMessage{}
	if err = json.Unmarshal(jsonData, &d); err != nil {
		s.handleHttpErr(c, err)
		return
	}

	id := uuid.NewV4().String()
	// TODO: try to add span to DB requests
	msg := porder.OrderMessage{Id: id, Name: d.Name, Amount: d.Amount, Status: "new"}
	_, err = s.coll.InsertOne(context.Background(), &msg)
	if err != nil {
		s.handleHttpErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, &msg)
}

func (s *service) handleProcessOrder(c *gin.Context) {
	id := c.Param("id")
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"status", "processed"}}}}
	// TODO: try to add span to DB requests
	_, err := s.coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		s.handleHttpErr(c, err)
		return
	}
	// TODO: try to add span to DB requests
	res := s.coll.FindOne(context.Background(), filter)
	if res.Err() != nil {
		s.handleHttpErr(c, res.Err())
		return
	}
	msg := porder.OrderMessage{}
	if err := res.Decode(&msg); err != nil {
		s.handleHttpErr(c, res.Err())
		return
	}

	c.JSON(http.StatusOK, &msg)
}

func (s *service) Listen() error {
	r := gin.Default()
	r.POST("/", s.handleCreateOrder)
	r.POST("/:id", s.handleProcessOrder)
	return r.Run()
}

func NewService(log *zap.Logger, mc metrics.Collector, coll *mongo.Collection, msgq mq.MQ) Service {
	bc := rest.NewBaseController(log, mc)
	return &service{BaseController: bc, coll: coll, mq: msgq}
}
