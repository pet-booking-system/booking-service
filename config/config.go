package config

import (
	"os"
)

type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	GRPCPort             string
	AuthServiceURL       string
	InventoryServiceAddr string
	PaymentServiceAddr   string
}

func Load() *Config {
	cfg := &Config{
		DBHost:               os.Getenv("DB_HOST"),
		DBPort:               os.Getenv("DB_PORT"),
		DBUser:               os.Getenv("DB_USER"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               os.Getenv("DB_NAME"),
		DBSSLMode:            os.Getenv("DB_SSLMODE"),
		GRPCPort:             os.Getenv("GRPC_PORT"),
		AuthServiceURL:       os.Getenv("AUTH_SERVICE_URL"),
		InventoryServiceAddr: os.Getenv("INVENTORY_SERVICE_ADDR"),
		PaymentServiceAddr:   os.Getenv("PAYMENT_SERVICE_ADDR"),
	}

	return cfg
}
