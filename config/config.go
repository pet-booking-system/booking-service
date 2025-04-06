package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost               string `mapstructure:"DB_HOST"`
	DBPort               string `mapstructure:"DB_PORT"`
	DBUser               string `mapstructure:"DB_USER"`
	DBPassword           string `mapstructure:"DB_PASSWORD"`
	DBName               string `mapstructure:"DB_NAME"`
	DBSSLMode            string `mapstructure:"DB_SSLMODE"`
	GRPCPort             string `mapstructure:"GRPC_PORT"`
	AuthServiceURL       string `mapstructure:"AUTH_SERVICE_URL"`
	InventoryServiceAddr string `mapstructure:"INVENTORY_SERVICE_ADDR"`
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found: %v (env variables will still be used)", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	return &cfg
}
