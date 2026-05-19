package balance

import (
	"context"
	"errors"

	"gophermart/internal/utils"
)

type Repository interface {
	GetBalance(ctx context.Context, userID int64) (Balance, error)
	CreateWithdrawal(ctx context.Context, userID int64, order string, sum float64) error
	ListWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Получает баланс пользователя
func (s *Service) GetBalance(ctx context.Context, userID int64) (Balance, error) {
	return s.repo.GetBalance(ctx, userID)
}

// Списывает средства с баланса пользователя
func (s *Service) Withdraw(ctx context.Context, userID int64, order string, sum float64) (int, error) {
	if !utils.ValidLuhn(order) {
		return 422, nil
	}
	if err := s.repo.CreateWithdrawal(ctx, userID, order, sum); err != nil {
		if errors.Is(err, ErrNotEnoughBalance) {
			return 402, nil
		}
		return 500, err
	}
	return 200, nil
}

// Получает список выводов средств
func (s *Service) ListWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error) {
	return s.repo.ListWithdrawals(ctx, userID)
}
