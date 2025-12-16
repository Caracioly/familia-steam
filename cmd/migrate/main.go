package main

import (
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mateus/familia-steam/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL não configurada")
	}

	database, err := db.Connect(databaseURL)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer database.Close()

	log.Println("Conectado ao banco. Aplicando migrations...")

	sqlBytes, err := os.ReadFile("migrations/001_init_schema.sql")
	if err != nil {
		log.Fatalf("Erro ao ler migration: %v", err)
	}

	if _, err := database.Exec(string(sqlBytes)); err != nil {
		log.Fatalf("Erro ao executar migration: %v", err)
	}

	log.Println("✓ Migrations aplicadas com sucesso!")
}
