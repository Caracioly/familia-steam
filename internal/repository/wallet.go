package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type Wallet struct {
	ID        int64
	UserID    int64
	CreatedAt time.Time
}

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) FindByUserID(userID int64) (*Wallet, error) {
	wallet := &Wallet{}
	err := r.db.QueryRow(`
		SELECT id, user_id, created_at
		FROM wallets
		WHERE user_id = $1
	`, userID).Scan(&wallet.ID, &wallet.UserID, &wallet.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar carteira: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) Create(userID int64) (*Wallet, error) {
	wallet := &Wallet{}
	err := r.db.QueryRow(`
		INSERT INTO wallets (user_id)
		VALUES ($1)
		RETURNING id, user_id, created_at
	`, userID).Scan(&wallet.ID, &wallet.UserID, &wallet.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar carteira: %w", err)
	}

	return wallet, nil
}

func (r *WalletRepository) FindOrCreate(userID int64) (*Wallet, error) {
	wallet, err := r.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	if wallet != nil {
		return wallet, nil
	}

	return r.Create(userID)
}

func (r *WalletRepository) GetBalance(walletID int64) (float64, error) {
	var balance sql.NullFloat64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE wallet_id = $1 AND status = 'CONFIRMED'
	`, walletID).Scan(&balance)

	if err != nil {
		return 0, fmt.Errorf("erro ao calcular saldo: %w", err)
	}

	return balance.Float64, nil
}

func (r *WalletRepository) GetTotalBalance() (float64, error) {
	var balance sql.NullFloat64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE status = 'CONFIRMED'
	`).Scan(&balance)

	if err != nil {
		return 0, fmt.Errorf("erro ao calcular saldo total: %w", err)
	}

	return balance.Float64, nil
}

type RankingEntry struct {
	Username string
	Balance  float64
}

func (r *WalletRepository) GetRanking(limit int) ([]RankingEntry, error) {
	rows, err := r.db.Query(`
		SELECT u.username, COALESCE(SUM(t.amount), 0) as balance
		FROM users u
		INNER JOIN wallets w ON w.user_id = u.id
		LEFT JOIN transactions t ON t.wallet_id = w.id AND t.status = 'CONFIRMED'
		GROUP BY u.id, u.username
		HAVING COALESCE(SUM(t.amount), 0) > 0
		ORDER BY balance DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar ranking: %w", err)
	}
	defer rows.Close()

	var ranking []RankingEntry
	for rows.Next() {
		var entry RankingEntry
		if err := rows.Scan(&entry.Username, &entry.Balance); err != nil {
			return nil, fmt.Errorf("erro ao ler ranking: %w", err)
		}
		ranking = append(ranking, entry)
	}

	return ranking, nil
}
