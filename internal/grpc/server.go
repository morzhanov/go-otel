package grpc

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type baseServer struct {
	//tracer opentracing.Tracer
	logger *zap.Logger
	url    string
}

type BaseServer interface {
	//PrepareContext(ctx context.Context) (context.Context, opentracing.Span)
	Listen(ctx context.Context, cancel context.CancelFunc, server *grpc.Server)
	Logger() *zap.Logger
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
		s.logger.Fatal("error during grpc server setup", zap.Error(err))
		return
	}

	if err := server.Serve(lis); err != nil {
		cancel()
		s.logger.Fatal("error during grpc server setup", zap.Error(err))
		return
	}
	s.logger.Info("Grpc server started", zap.String("port", s.url))
	<-ctx.Done()
	if err := lis.Close(); err != nil {
		cancel()
		s.logger.Fatal("error during grpc server setup", zap.Error(err))
		return
	}
}

func (s *baseServer) Logger() *zap.Logger {
	return s.logger
}

func NewServer(
	//tracer opentracing.Tracer,
	logger *zap.Logger,
	url string,
) BaseServer {
	return &baseServer{logger, url}
}
