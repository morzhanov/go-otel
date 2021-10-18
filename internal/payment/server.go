package payment

import (
	"context"
	"fmt"
	"net"

	gpayment "github.com/morzhanov/go-otel/api/grpc/payment"
	gserver "github.com/morzhanov/go-otel/internal/grpc/server"
	"github.com/morzhanov/go-otel/internal/mq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	gpayment.UnimplementedPaymentServer
	gserver.BaseServer
	mq  mq.MQ
	srv *grpc.Server
	url string
	pay Payment
}

type Server interface {
	//PrepareContext(ctx context.Context) (context.Context, opentracing.Span)
	Listen(ctx context.Context, cancel context.CancelFunc, server *grpc.Server)
}

//func (s *server) PrepareContext(ctx context.Context) (context.Context, opentracing.Span) {
//	span := tracing.StartSpanFromGrpcRequest(s.Tracer, ctx)
//	ctx = context.WithValue(ctx, "transport", sender.RpcTransport)
//	return ctx, span
//}

func (s *server) GetPaymentInfo(ctx context.Context, in *gpayment.GetPaymentInfoRequest) (*gpayment.PaymentMessage, error) {
	ctx, span := s.PrepareContext(ctx)
	defer span.Finish()
	return s.pay.GetPaymentInfo(in)
}

func (s *server) Listen(ctx context.Context, cancel context.CancelFunc, srv *grpc.Server) {
	lis, err := net.Listen("tcp", s.url)
	if err != nil {
		cancel()
		s.BaseServer.Logger().Fatal("error during grpc service start")
		return
	}

	if err := srv.Serve(lis); err != nil {
		cancel()
		s.BaseServer.Logger().Fatal("error during grpc service start")
		return
	}
	s.BaseServer.Logger().Info("Grpc srv started", zap.String("port", s.url))
	<-ctx.Done()
	if err := lis.Close(); err != nil {
		cancel()
		s.BaseServer.Logger().Fatal("error during grpc service start")
		return
	}
}

func NewServer(
	//tracer opentracing.Tracer,
	grpcAddr string,
	grpcPort string,
	logger *zap.Logger,
	pay Payment,
) Server {
	url := fmt.Sprintf("%s:%s", grpcAddr, grpcPort)
	bs := gserver.NewServer(logger, url)
	s := &server{BaseServer: bs, srv: grpc.NewServer(), url: url, pay: pay}
	gpayment.RegisterPaymentServer(s.srv, s)
	reflection.Register(s.srv)
	return s
}
