package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/mateus/familia-steam/internal/api"
	"github.com/mateus/familia-steam/internal/bot"
	"github.com/mateus/familia-steam/internal/config"
	"github.com/mateus/familia-steam/internal/db"
	"github.com/mateus/familia-steam/internal/mercadopago"
	"github.com/mateus/familia-steam/internal/repository"
	"github.com/mateus/familia-steam/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer database.Close()
	log.Println("Conectado ao banco de dados")

	userRepo := repository.NewUserRepository(database)
	walletRepo := repository.NewWalletRepository(database)
	txRepo := repository.NewTransactionRepository(database)

	mpClient := mercadopago.NewClient(cfg.MercadoPagoToken)

	paymentService := service.NewPaymentService(mpClient, txRepo, userRepo, walletRepo)
	walletService := service.NewWalletService(userRepo, walletRepo)

	apiURL := fmt.Sprintf("http://localhost:%s", cfg.Port)

	discordBot, err := bot.New(cfg.DiscordToken, apiURL)
	if err != nil {
		log.Fatalf("Erro ao criar bot do Discord: %v", err)
	}

	if err := discordBot.Start(); err != nil {
		log.Fatalf("Erro ao iniciar bot do Discord: %v", err)
	}
	defer discordBot.Stop()

	server := api.New(cfg.Port, database, paymentService, walletService)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Erro no servidor HTTP: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Iniciando shutdown gracioso...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Erro ao encerrar servidor HTTP: %v", err)
	}

	log.Println("Aplicação encerrada")
}
