package grpc

import (
	"context"
	"net"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/win5do/golang-microservice-demo/pkg/log"

	"github.com/win5do/golang-microservice-demo/pkg/api/petpb"
	"github.com/win5do/golang-microservice-demo/pkg/config/util"
	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
	petdb "github.com/win5do/golang-microservice-demo/pkg/repository/db/pet"
	petsvc "github.com/win5do/golang-microservice-demo/pkg/service/pet"

	"github.com/win5do/golang-microservice-demo/pkg/config"
)

func Run(ctx context.Context, cfg *config.Config) {
	wg := util.GetWaitGroupInCtx(ctx)
	wg.Add(1)
	defer wg.Done()

	addr := net.JoinHostPort("", cfg.GrpcPort)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	logger := log.GetLogger()
	grpc_zap.ReplaceGrpcLogger(logger)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	petpb.RegisterPetServiceServer(s, petsvc.NewPetService(dbcore.NewTxImpl(), petdb.NewPetDomain()))

	go func() {
		// Run the server
		log.Infof("grpc server start: %s", addr)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		if err := runGateway(net.JoinHostPort("", cfg.GrpcGatewayPort), addr); err != nil {
			log.Fatalf("err: %+v", err)
		}
	}()

	<-ctx.Done()
}
