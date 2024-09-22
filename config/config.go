package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Config holds the configuration for the email service.
// It includes nested configurations for the service itself and email settings.
//
// Fields:
//   - Service: The Service struct containing the service-related configuration.
//   - Email: The Email struct containing the email-related configuration.
type Config struct {
	Service Service
	Email   Email
}

// Service holds the configuration for the service, including the port and listen address.
//
// Fields:
//   - Port: The port on which the email service will listen. It is loaded from the environment variable "EMAIL_SERVICE_PORT" with a default value of 8080.
//   - ListenAddress: The address on which the email service will listen. It is loaded from the environment variable "EMAIL_SERVICE_LISTEN_ADDRESS" with a default value of "0.0.0.0".
type Service struct {
	ListenAddress string `env:"EMAIL_SERVICE_LISTEN_ADDRESS" envDefault:"0.0.0.0"`
	Port          int    `env:"EMAIL_SERVICE_PORT" envDefault:"8080"`
	GRPCHost      string `env:"EMAIL_SERVICE_GRPC_HOST" envDefault:"127.0.0.1"`
	GRPCPort      int    `env:"EMAIL_SERVICE_GRPC_PORT" envDefault:"8081"`
}

// Email holds the configuration for email settings, including the sender and forward addresses.
//
// Fields:
//   - From: The email address from which emails will be sent. It is loaded from the environment variable "EMAIL_SERVICE_EMAIL_FROM".
//   - Forward: The email address to which incoming emails will be forwarded. It is loaded from the environment variable "EMAIL_SERVICE_EMAIL_FORWARD".
//   - ThankYouTemplate: A base64 standard encoded html template for your thank you email.
type Email struct {
	From             string `env:"EMAIL_SERVICE_EMAIL_FROM"`
	Forward          string `env:"EMAIL_SERVICE_EMAIL_FORWARD"`
	ThankYouTemplate string `env:"EMAIL_SERVICE_EMAIL_THANK_YOU_TEMPLATE"`
}

// Load loads the configuration from environment variables using the env package.
// It returns a pointer to the Config struct and an error if any occurred during the loading process.
//
// Returns:
//   - *Config: The loaded configuration.
//   - error: An error if any occurred during the loading of the environment variables.
func Load() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return &cfg, fmt.Errorf("failed to load environment: %s", err.Error())
	}

	return &cfg, nil
}
