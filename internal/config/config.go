package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port             string
	DatabaseURL      string
	DiscordToken     string
	MercadoPagoToken string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL é obrigatória")
	}

	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN é obrigatória")
	}

	mercadoPagoToken := os.Getenv("MERCADOPAGO_ACCESS_TOKEN")
	if mercadoPagoToken == "" {
		return nil, fmt.Errorf("MERCADOPAGO_ACCESS_TOKEN é obrigatória")
	}

	return &Config{
		Port:             port,
		DatabaseURL:      databaseURL,
		DiscordToken:     discordToken,
		MercadoPagoToken: mercadoPagoToken,
	}, nil
}
