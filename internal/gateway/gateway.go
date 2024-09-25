package gateway

import (
	"context"
	"fmt"
	"net/http"

	mailservice_v1 "github.com/brice-aldrich/mail-service/gen/go/mailservice.v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

// Config holds the configuration for the gRPC-Gateway server.
// It includes the host and port for both the HTTP server and the gRPC server.
//
// Fields:
//   - Host: The host address for the HTTP server.
//   - Port: The port number for the HTTP server.
//   - GRPCHost: The host address for the gRPC server.
//   - GRPCPort: The port number for the gRPC server.
type Config struct {
	Host     string
	Port     int
	GRPCHost string
	GRPCPort int
}

// gateway represents the gRPC-Gateway server.
// It includes the host and port for both the HTTP server and the gRPC server, as well as the ServeMux for routing HTTP requests.
//
// Fields:
//   - host: The host address for the HTTP server.
//   - port: The port number for the HTTP server.
//   - grpcHost: The host address for the gRPC server.
//   - grpcPort: The port number for the gRPC server.
//   - mux: The runtime.ServeMux for routing HTTP requests to gRPC handlers.
type gateway struct {
	host     string
	port     int
	grpcHost string
	grpcPort int
	mux      *runtime.ServeMux
}

// New creates a new instance of the gateway with the provided configuration.
// It initializes the ServeMux for routing HTTP requests.
//
// Parameters:
//   - cfg: The Config object containing the host and port for both the HTTP server and the gRPC server.
//
// Returns:
//   - *gateway: The newly created gateway instance.
func New(cfg Config) *gateway {
	return &gateway{
		host:     cfg.Host,
		port:     cfg.Port,
		grpcHost: cfg.GRPCHost,
		grpcPort: cfg.GRPCPort,
		mux:      runtime.NewServeMux(),
	}
}

// Register registers the MailService handler with the gRPC-Gateway mux.
// It connects the mux to the gRPC server endpoint.
//
// Parameters:
//   - ctx: The context.Context object for the request.
//   - opts: Additional grpc.DialOption options for the gRPC connection.
//
// Returns:
//   - error: An error if any occurred during the registration of the handler.
func (g gateway) Register(ctx context.Context, opts ...grpc.DialOption) error {
	return mailservice_v1.RegisterMailServiceHandlerFromEndpoint(ctx, g.mux, fmt.Sprintf("%s:%d", g.grpcHost, g.grpcPort), opts)
}

// Serve starts the HTTP server and listens for incoming requests.
// It applies CORS settings to allow cross-origin requests.
//
// Returns:
//   - error: An error if any occurred during the server startup or while listening for requests.
func (g gateway) Serve() error {
	withCors := cors.New(cors.Options{
		AllowedOrigins: []string{"https://www.bricealdrich.com", "http://localhost:3000"},
		AllowedMethods: []string{http.MethodPost, http.MethodOptions, http.MethodGet},
	}).Handler(g.mux)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", g.host, g.port),
		Handler: withCors,
	}

	return server.ListenAndServe()
}
