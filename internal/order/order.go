package order

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/morzhanov/go-otel/internal/order/model"
)

type service struct {
}

type Service interface {
	Listen() error
}

func handleHttpErr(c *gin.Context, err error) {
	c.String(http.StatusInternalServerError, err.Error())
	log.Println(fmt.Errorf("error in the handler: %w", err))
}

func (o *service) handleAddOrder(c *gin.Context) {
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {

		return
	}
	order := model.Order{}
	if err = json.Unmarshal(jsonData, &order); err != nil {
		handleHttpErr(c, err)
		return
	}
	if err := o.cs.Add("create_order", &order); err != nil {
		handleHttpErr(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (o *service) handleProcessOrder(c *gin.Context) {
	id := c.Param("id")
	if err := o.cs.Add("process_order", id); err != nil {
		handleHttpErr(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (o *service) handleGetOrders(c *gin.Context) {
	res, err := o.qs.GetAll()
	if err != nil {
		handleHttpErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (o *service) handleGetOrder(c *gin.Context) {
	id := c.Param("id")
	res, err := o.qs.Get(id)
	if err != nil {
		handleHttpErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (o *service) Listen() error {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AddAllowHeaders([]string{"authorization"}...)
	router.Use(cors.New(config))
	r := gin.Default()
	r.POST("/", o.handleAddOrder)
	r.POST("/:id", o.handleProcessOrder)
	r.GET("/", o.handleGetOrders)
	r.GET("/:id", o.handleGetOrder)
	return r.Run()
}

func NewService(cs internal.CommandStore, qs internal.QueryStore) Service {
	return &service{cs, qs}
}
