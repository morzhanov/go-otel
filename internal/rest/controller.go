package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/morzhanov/go-otel/internal/telemetry/meter"

	"go.opentelemetry.io/otel/trace"

	"github.com/morzhanov/go-otel/internal/telemetry"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type baseController struct {
	router *gin.Engine
	log    *zap.Logger
	tel    telemetry.Telemetry
}

type BaseController interface {
	Listen(port string)
	ParseRestBody(ctx *gin.Context, input interface{}) error
	HandleRestError(ctx *gin.Context, err error)
	Handler(handler gin.HandlerFunc) gin.HandlerFunc
	Router() *gin.Engine
	Logger() *zap.Logger
	Tracer() telemetry.TraceFn
	Meter() meter.Meter
}

func (c *baseController) Listen(
	port string,
) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: c.router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.log.Fatal("error during REST controller setup", zap.Error(err))
			return
		}
	}()
	c.log.Info("gRPC server started")
	ctx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	if err := srv.Shutdown(ctx); err != nil {
		cancel2()
		c.log.Fatal("error during REST controller setup", zap.Error(err))
	}
}

func (c *baseController) ParseRestBody(ctx *gin.Context, input interface{}) error {
	jsonData, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}

	in := reflect.ValueOf(input).Interface()
	return json.Unmarshal(jsonData, &in)
}

func (c *baseController) HandleRestError(ctx *gin.Context, err error) {
	c.log.Error(errors.Unwrap(err).Error())
	if err.Error() == "not authorized" {
		ctx.String(http.StatusUnauthorized, err.Error())
		return
	}
	ctx.String(http.StatusInternalServerError, err.Error())
}

func (c *baseController) Handler(handler gin.HandlerFunc) gin.HandlerFunc {
	return handler
}

func PerformRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	sci := ctx.Value("span-context")
	sc := sci.(trace.SpanContext)
	req.Header.Set("content-type", "application/json")
	spanCtx, err := sc.MarshalJSON()
	if err != nil {
		return nil, err
	}
	req.Header.Set("span-context", string(spanCtx))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

func GetSpanContext(ctx *gin.Context) (*context.Context, error) {
	scs := ctx.GetHeader("span-context")
	var sctx context.Context
	if err := json.Unmarshal([]byte(scs), &sctx); err != nil {
		return nil, err
	}
	return &sctx, nil
}

func (c *baseController) Router() *gin.Engine       { return c.router }
func (c *baseController) Logger() *zap.Logger       { return c.log }
func (c *baseController) Tracer() telemetry.TraceFn { return c.tel.Tracer() }
func (c *baseController) Meter() meter.Meter        { return c.tel.Meter() }

func NewBaseController(log *zap.Logger, tel telemetry.Telemetry) BaseController {
	router := gin.Default()
	return &baseController{router: router, log: log, tel: tel}
}
