package mercadopago

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Client struct {
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		baseURL:     "https://api.mercadopago.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type PixPaymentRequest struct {
	TransactionAmount float64 `json:"transaction_amount"`
	Description       string  `json:"description"`
	PaymentMethodID   string  `json:"payment_method_id"`
	Payer             Payer   `json:"payer"`
}

type Payer struct {
	Email string `json:"email"`
}

type PixPaymentResponse struct {
	ID                 int64              `json:"id"`
	Status             string             `json:"status"`
	StatusDetail       string             `json:"status_detail"`
	TransactionAmount  float64            `json:"transaction_amount"`
	PointOfInteraction PointOfInteraction `json:"point_of_interaction"`
}

type PointOfInteraction struct {
	TransactionData TransactionData `json:"transaction_data"`
}

type TransactionData struct {
	QRCode       string `json:"qr_code"`
	QRCodeBase64 string `json:"qr_code_base64"`
	TicketURL    string `json:"ticket_url"`
}

func (c *Client) CreatePixPayment(amount float64, description string) (*PixPaymentResponse, error) {
	reqBody := PixPaymentRequest{
		TransactionAmount: amount,
		Description:       description,
		PaymentMethodID:   "pix",
		Payer: Payer{
			Email: "payer@email.com",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/v1/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("X-Idempotency-Key", generateIdempotencyKey())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API Mercado Pago [%d]: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Mercado Pago Response [%d]: %s\n", resp.StatusCode, string(body))

	var payment PixPaymentResponse
	if err := json.Unmarshal(body, &payment); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	fmt.Printf("Parsed Payment: ID=%d, Status=%s, QRCode=%s, QRCodeBase64=%s\n",
		payment.ID, payment.Status,
		payment.PointOfInteraction.TransactionData.QRCode[:min(50, len(payment.PointOfInteraction.TransactionData.QRCode))],
		payment.PointOfInteraction.TransactionData.QRCodeBase64[:min(50, len(payment.PointOfInteraction.TransactionData.QRCodeBase64))])

	return &payment, nil
}

func (c *Client) GetPayment(paymentID int64) (*PixPaymentResponse, error) {
	url := fmt.Sprintf("%s/v1/payments/%d", c.baseURL, paymentID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na API Mercado Pago [%d]: %s", resp.StatusCode, string(body))
	}

	var payment PixPaymentResponse
	if err := json.Unmarshal(body, &payment); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return &payment, nil
}

func generateIdempotencyKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
