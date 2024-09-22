package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/brice-aldrich/mail-service/config"
	mailservice_v1 "github.com/brice-aldrich/mail-service/gen/go/mailservice.v1"
	"github.com/brice-aldrich/mail-service/internal/gateway"
	"github.com/brice-aldrich/mail-service/internal/mail"
	"github.com/brice-aldrich/mail-service/internal/server"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	zlog, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to start logger: %s", err.Error())
	}

	cfg, err := config.Load()
	if err != nil {
		zlog.With(zap.Error(err)).Fatal("Failed to load application configuration.")
	}

	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion("us-east-1"))
	if err != nil {
		zlog.With(zap.Error(err)).Fatal("Failed to load AWS configuration.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	mailOrch, err := mail.New(ctx, mail.Config{
		SES:          sesv2.NewFromConfig(awsConfig),
		ForwardEmail: cfg.Email.Forward,
		FromEmail:    cfg.Email.From,
	})
	if err != nil {
		zlog.With(zap.Error(err)).Fatal("Failed to setup mail orchestrator.")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zap.UnaryServerInterceptor(zlog),
		),
	)

	mailService := server.New(mailOrch)
	mailservice_v1.RegisterMailServiceServer(grpcServer, mailService)

	gw := gateway.New(gateway.Config{
		Host:     cfg.Service.ListenAddress,
		Port:     cfg.Service.Port,
		GRPCHost: cfg.Service.GRPCHost,
		GRPCPort: cfg.Service.GRPCPort,
	})

	if err := gw.Register(context.Background(), grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		zlog.With(zap.Error(err)).Fatal("Failed to register gRPC gateway.")
	}

	go func() {
		if err := gw.Serve(); err != nil {
			zlog.With(zap.Error(err)).Fatal("Failed to service gRPC gateway.")
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Service.GRPCHost, cfg.Service.GRPCPort))
	if err != nil {
		zlog.With(zap.Error(err), zap.Int("port", cfg.Service.Port), zap.String("host", cfg.Service.ListenAddress)).Fatal("Failed to open TCP socket.")
	}

	zlog.With(zap.Int("port", cfg.Service.Port), zap.String("host", cfg.Service.ListenAddress)).Info("Starting Email Service.")
	if err := grpcServer.Serve(lis); err != nil {
		zlog.With(zap.Error(err), zap.Int("port", cfg.Service.Port), zap.String("host", cfg.Service.ListenAddress)).Fatal("Failed to start email service.")
	}
}
