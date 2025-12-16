package service

import (
	"fmt"

	"github.com/mateus/familia-steam/internal/repository"
)

type WalletService struct {
	userRepo   *repository.UserRepository
	walletRepo *repository.WalletRepository
}

func NewWalletService(
	userRepo *repository.UserRepository,
	walletRepo *repository.WalletRepository,
) *WalletService {
	return &WalletService{
		userRepo:   userRepo,
		walletRepo: walletRepo,
	}
}

func (s *WalletService) GetUserBalance(discordID string) (float64, error) {
	user, err := s.userRepo.FindByDiscordID(discordID)
	if err != nil {
		return 0, fmt.Errorf("erro ao buscar usuário: %w", err)
	}
	if user == nil {
		return 0, nil
	}

	wallet, err := s.walletRepo.FindByUserID(user.ID)
	if err != nil {
		return 0, fmt.Errorf("erro ao buscar carteira: %w", err)
	}
	if wallet == nil {
		return 0, nil
	}

	balance, err := s.walletRepo.GetBalance(wallet.ID)
	if err != nil {
		return 0, fmt.Errorf("erro ao calcular saldo: %w", err)
	}

	return balance, nil
}

func (s *WalletService) GetTotalBalance() (float64, error) {
	balance, err := s.walletRepo.GetTotalBalance()
	if err != nil {
		return 0, fmt.Errorf("erro ao calcular saldo total: %w", err)
	}
	return balance, nil
}

func (s *WalletService) GetRanking(limit int) ([]repository.RankingEntry, error) {
	ranking, err := s.walletRepo.GetRanking(limit)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar ranking: %w", err)
	}
	return ranking, nil
}
