package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/brice-aldrich/mail-service/config"
	"github.com/brice-aldrich/mail-service/internal/mail"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
type contextKey string

const (
	correlationIDKey contextKey = "X-Correlation-ID"
	loggerKey        contextKey = "logger"
)

var (
	cfg      *config.Config
	mailOrch mail.Orchestrator
	zlog     *zap.Logger
)

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Unable to initialize logger, %v", err)
	}
	zlog = logger

	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		zlog.With(zap.Error(err)).Fatal("Failed to load AWS configuration.")
	}

	cfg, err = config.Load()
	if err != nil {
		zlog.With(zap.Error(err)).Fatal("Failed to load app config.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	mailOrch, err = mail.New(ctx, mail.Config{
		SES:    sesv2.NewFromConfig(awsCfg),
		Logger: zlog,
		Cfg:    cfg,
	})
	if err != nil {
		zlog.With(zap.Error(err)).Fatal("Failed to setup mail orchestrator.")
	}
}

func withRequestLogging(logger *zap.Logger) func(Handler) Handler {
	return func(next Handler) Handler {
		return func(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
			startTime := time.Now()

			correlationID := request.Headers["X-Correlation-ID"]
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			requestLogger := logger.With(
				zap.String("correlation_id", correlationID),
				zap.String("request_id", request.RequestContext.RequestID),
				zap.String("path", request.Path),
				zap.String("method", request.HTTPMethod),
			)

			ctx = context.WithValue(ctx, correlationIDKey, correlationID)
			ctx = context.WithValue(ctx, loggerKey, requestLogger)

			requestLogger.Info("Processing request",
				zap.Any("query_params", request.QueryStringParameters),
				zap.Any("path_params", request.PathParameters),
			)

			response, err = next(ctx, request)
			duration := time.Since(startTime)

			if err != nil {
				requestLogger.Error("Request Failed",
					zap.Error(err),
					zap.Duration("duration_ms", duration),
				)
			} else {
				requestLogger.Info("Request Completed",
					zap.Int("status_code", response.StatusCode),
					zap.Duration("duration_ms", duration),
				)
			}

			if response.Headers == nil {
				response.Headers = make(map[string]string)
			}
			response.Headers["X-Correlation-ID"] = correlationID

			return response, err
		}
	}
}

func correlationIDFromContext(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cid := correlationIDFromContext(ctx)

	if err := mailOrch.SendMail(ctx, json.RawMessage(request.Body)); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("failed to process request with x-correlation-id: %s", cid),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Success",
	}, nil
}

func main() {
	handler := withRequestLogging(zlog)(handleRequest)
	lambda.Start(handler)
}
