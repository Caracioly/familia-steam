package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "PENDING"
	StatusConfirmed TransactionStatus = "CONFIRMED"
	StatusFailed    TransactionStatus = "FAILED"
)

type Transaction struct {
	ID                int64
	WalletID          int64
	Amount            float64
	Status            TransactionStatus
	ExternalReference string
	PaymentData       map[string]interface{}
	CreatedAt         time.Time
	ConfirmedAt       *time.Time
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(walletID int64, amount float64, externalRef string, paymentData map[string]interface{}) (*Transaction, error) {
	paymentJSON, err := json.Marshal(paymentData)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payment_data: %w", err)
	}

	tx := &Transaction{}
	err = r.db.QueryRow(`
		INSERT INTO transactions (wallet_id, amount, status, external_reference, payment_data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, wallet_id, amount, status, external_reference, payment_data, created_at, confirmed_at
	`, walletID, amount, StatusPending, externalRef, paymentJSON).Scan(
		&tx.ID, &tx.WalletID, &tx.Amount, &tx.Status,
		&tx.ExternalReference, &paymentJSON, &tx.CreatedAt, &tx.ConfirmedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar transação: %w", err)
	}

	json.Unmarshal(paymentJSON, &tx.PaymentData)
	return tx, nil
}

func (r *TransactionRepository) FindByExternalReference(externalRef string) (*Transaction, error) {
	tx := &Transaction{}
	var paymentJSON []byte

	err := r.db.QueryRow(`
		SELECT id, wallet_id, amount, status, external_reference, payment_data, created_at, confirmed_at
		FROM transactions
		WHERE external_reference = $1
	`, externalRef).Scan(
		&tx.ID, &tx.WalletID, &tx.Amount, &tx.Status,
		&tx.ExternalReference, &paymentJSON, &tx.CreatedAt, &tx.ConfirmedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}

	json.Unmarshal(paymentJSON, &tx.PaymentData)
	return tx, nil
}

func (r *TransactionRepository) UpdateStatus(id int64, status TransactionStatus) error {
	var confirmedAt interface{}
	if status == StatusConfirmed {
		confirmedAt = time.Now()
	}

	result, err := r.db.Exec(`
		UPDATE transactions
		SET status = $1, confirmed_at = $2
		WHERE id = $3 AND status != $1
	`, status, confirmedAt, id)

	if err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil
	}

	return nil
}
