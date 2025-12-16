package service

import (
	"fmt"

	"github.com/mateus/familia-steam/internal/mercadopago"
	"github.com/mateus/familia-steam/internal/repository"
)

type PaymentService struct {
	mpClient   *mercadopago.Client
	txRepo     *repository.TransactionRepository
	userRepo   *repository.UserRepository
	walletRepo *repository.WalletRepository
}

func NewPaymentService(
	mpClient *mercadopago.Client,
	txRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	walletRepo *repository.WalletRepository,
) *PaymentService {
	return &PaymentService{
		mpClient:   mpClient,
		txRepo:     txRepo,
		userRepo:   userRepo,
		walletRepo: walletRepo,
	}
}

type CreatePixPaymentRequest struct {
	DiscordID string
	Username  string
	Amount    float64
}

type CreatePixPaymentResponse struct {
	TransactionID     int64   `json:"transaction_id"`
	Amount            float64 `json:"amount"`
	QRCode            string  `json:"qr_code"`
	QRCodeBase64      string  `json:"qr_code_base64"`
	ExternalReference string  `json:"external_reference"`
}

func (s *PaymentService) CreatePixPayment(req CreatePixPaymentRequest) (*CreatePixPaymentResponse, error) {
	user, err := s.userRepo.FindOrCreate(req.DiscordID, req.Username)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar/criar usuário: %w", err)
	}

	wallet, err := s.walletRepo.FindOrCreate(user.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar/criar carteira: %w", err)
	}

	description := fmt.Sprintf("Vaquinha - %s - R$ %.2f", req.Username, req.Amount)
	payment, err := s.mpClient.CreatePixPayment(req.Amount, description)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pagamento: %w", err)
	}

	if payment.ID == 0 {
		return nil, fmt.Errorf("mercado pago retornou ID inválido - verifique o token (use APP_USR-... para produção)")
	}

	externalRef := fmt.Sprintf("%d", payment.ID)
	paymentData := map[string]interface{}{
		"qr_code":        payment.PointOfInteraction.TransactionData.QRCode,
		"qr_code_base64": payment.PointOfInteraction.TransactionData.QRCodeBase64,
		"ticket_url":     payment.PointOfInteraction.TransactionData.TicketURL,
		"status":         payment.Status,
	}

	transaction, err := s.txRepo.Create(wallet.ID, req.Amount, externalRef, paymentData)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar transação: %w", err)
	}

	return &CreatePixPaymentResponse{
		TransactionID:     transaction.ID,
		Amount:            transaction.Amount,
		QRCode:            payment.PointOfInteraction.TransactionData.QRCode,
		QRCodeBase64:      payment.PointOfInteraction.TransactionData.QRCodeBase64,
		ExternalReference: externalRef,
	}, nil
}

type PaymentConfirmedData struct {
	DiscordID string
	Username  string
	Amount    float64
}

func (s *PaymentService) ConfirmPayment(externalRef string) (*PaymentConfirmedData, error) {
	transaction, err := s.txRepo.FindByExternalReference(externalRef)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}
	if transaction == nil {
		return nil, fmt.Errorf("transação não encontrada: %s", externalRef)
	}

	if err := s.txRepo.UpdateStatus(transaction.ID, repository.StatusConfirmed); err != nil {
		return nil, fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	wallet, err := s.walletRepo.FindByID(transaction.WalletID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar wallet: %w", err)
	}

	user, err := s.userRepo.FindByID(wallet.UserID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	return &PaymentConfirmedData{
		DiscordID: user.DiscordID,
		Username:  user.Username,
		Amount:    transaction.Amount,
	}, nil
}
