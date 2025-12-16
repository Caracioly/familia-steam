package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mateus/familia-steam/internal/service"
)

type Server struct {
	server         *http.Server
	db             *sql.DB
	paymentService *service.PaymentService
	walletService  *service.WalletService
}

func New(port string, db *sql.DB, paymentService *service.PaymentService, walletService *service.WalletService) *Server {
	mux := http.NewServeMux()

	s := &Server{
		server: &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		db:             db,
		paymentService: paymentService,
		walletService:  walletService,
	}

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/payments/create", s.handleCreatePayment)
	mux.HandleFunc("/api/payments/webhook", s.handleWebhook)
	mux.HandleFunc("/api/wallet/balance", s.handleGetBalance)
	mux.HandleFunc("/api/wallet/ranking", s.handleGetRanking)

	return s
}

func (s *Server) Start() error {
	log.Printf("Servidor HTTP iniciado na porta %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("erro ao iniciar servidor HTTP: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Encerrando servidor HTTP...")
	return s.server.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database unavailable"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Família Steam API"))
}

func (s *Server) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DiscordID string  `json:"discord_id"`
		Username  string  `json:"username"`
		Amount    float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Valor deve ser maior que zero", http.StatusBadRequest)
		return
	}

	payment, err := s.paymentService.CreatePixPayment(service.CreatePixPaymentRequest{
		DiscordID: req.DiscordID,
		Username:  req.Username,
		Amount:    req.Amount,
	})

	if err != nil {
		log.Printf("Erro ao criar pagamento: %v", err)
		http.Error(w, "Erro ao criar pagamento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var webhook struct {
		Action string `json:"action"`
		Data   struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		log.Printf("Erro ao decodificar webhook: %v", err)
		http.Error(w, "Requisição inválida", http.StatusBadRequest)
		return
	}

	if webhook.Action == "payment.updated" || webhook.Action == "payment.created" {
		if webhook.Data.ID != "" {
			confirmed, err := s.paymentService.ConfirmPayment(webhook.Data.ID)
			if err != nil {
				log.Printf("Erro ao confirmar pagamento %s: %v", webhook.Data.ID, err)
			} else if confirmed != nil {
				log.Printf("✅ Pagamento confirmado! User: %s, Valor: R$ %.2f", confirmed.Username, confirmed.Amount)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	discordID := r.URL.Query().Get("discord_id")
	if discordID == "" {
		http.Error(w, "discord_id é obrigatório", http.StatusBadRequest)
		return
	}

	balance, err := s.walletService.GetUserBalance(discordID)
	if err != nil {
		log.Printf("Erro ao buscar saldo: %v", err)
		http.Error(w, "Erro ao buscar saldo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}

func (s *Server) handleGetRanking(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ranking, err := s.walletService.GetRanking(limit)
	if err != nil {
		log.Printf("Erro ao buscar ranking: %v", err)
		http.Error(w, "Erro ao buscar ranking", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ranking)
}
