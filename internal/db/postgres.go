package db

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/lib/pq"
)

// Connect estabelece a conexão com o banco de dados PostgreSQL
func Connect(databaseURL string) (*sql.DB, error) {
	// Parse da URL para garantir que sslmode=require está presente
	u, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear DATABASE_URL: %w", err)
	}

	// Adiciona sslmode=require se não estiver presente
	q := u.Query()
	if q.Get("sslmode") == "" {
		q.Set("sslmode", "require")
		u.RawQuery = q.Encode()
	}

	// Conecta ao banco
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com o banco: %w", err)
	}

	// Configura o pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Verifica a conexão
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao conectar ao banco: %w", err)
	}

	return db, nil
}
