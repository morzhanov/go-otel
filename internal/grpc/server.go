package grpc

import (
	"context"
	"net"

	"github.com/morzhanov/go-otel/internal/telemetry"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type baseServer struct {
	url string
	log *zap.Logger
	tel telemetry.Telemetry
}

type BaseServer interface {
	Listen(ctx context.Context, cancel context.CancelFunc, server *grpc.Server)
	Logger() *zap.Logger
	Tracer() telemetry.TraceFn
	Meter() metric.Meter
	//PrepareContext(ctx context.Context) (context.Context, opentracing.Span)
}

//func (s *baseServer) PrepareContext(ctx context.Context) (context.Context, opentracing.Span) {
//	span := tracing.StartSpanFromGrpcRequest(s.Tracer, ctx)
//	ctx = context.WithValue(ctx, "transport", sender.RpcTransport)
//	return ctx, span
//}

func (s *baseServer) Listen(ctx context.Context, cancel context.CancelFunc, server *grpc.Server) {
	lis, err := net.Listen("tcp", s.url)
	if err != nil {
		cancel()
		s.log.Fatal("error during grpc server setup", zap.Error(err))
		return
	}

	if err := server.Serve(lis); err != nil {
		cancel()
		s.log.Fatal("error during grpc server setup", zap.Error(err))
		return
	}
	s.log.Info("Grpc server started", zap.String("port", s.url))
	<-ctx.Done()
	if err := lis.Close(); err != nil {
		cancel()
		s.log.Fatal("error during grpc server setup", zap.Error(err))
		return
	}
}

func (s *baseServer) Logger() *zap.Logger       { return s.log }
func (s *baseServer) Tracer() telemetry.TraceFn { return s.tel.Tracer() }
func (s *baseServer) Meter() metric.Meter       { return s.tel.Meter() }

func NewServer(url string, log *zap.Logger, tel telemetry.Telemetry) BaseServer {
	return &baseServer{log: log, url: url, tel: tel}
}
