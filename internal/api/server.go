package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Server representa o servidor HTTP
type Server struct {
	server *http.Server
	db     *sql.DB
}

// New cria uma nova instância do servidor
func New(port string, db *sql.DB) *Server {
	mux := http.NewServeMux()

	s := &Server{
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		db: db,
	}

	// Registra rotas
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/", s.handleRoot)

	return s
}

// Start inicia o servidor HTTP (bloqueante)
func (s *Server) Start() error {
	log.Printf("Servidor HTTP iniciado na porta %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("erro ao iniciar servidor HTTP: %w", err)
	}
	return nil
}

// Shutdown encerra o servidor graciosamente
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Encerrando servidor HTTP...")
	return s.server.Shutdown(ctx)
}

// handleHealth retorna o status do servidor
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Verifica a conexão com o banco
	if err := s.db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database unavailable"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleRoot é a rota raiz
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Família Steam API"))
}
