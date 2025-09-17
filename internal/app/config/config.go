package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCAddr       string `envconfig:"GRPC_ADDR" required:"true"`
	HTTPAddr       string `envconfig:"HTTP_ADDR" required:"true"`
	DSN            string `envconfig:"DSN"  required:"true"`
	MigrationsPath string `envconfig:"MIGRATIONS_PATH" required:"true"`
}

// Read reads config from environment using envconfig.
func Read() (Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
