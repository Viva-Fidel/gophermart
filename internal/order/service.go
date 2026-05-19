package order

import (
	"context"

	"gophermart/internal/utils"
)

type Repository interface {
	CreateOrder(ctx context.Context, userID int64, number string) (created bool, ownerID int64, err error)
	ListOrders(ctx context.Context, userID int64) ([]Order, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Загружает заказ
func (s *Service) SubmitOrder(ctx context.Context, userID int64, number string) (int, error) {
	if !utils.ValidLuhn(number) {
		return 422, nil
	}
	created, ownerID, err := s.repo.CreateOrder(ctx, userID, number)
	if err != nil {
		return 500, err
	}
	if created {
		return 202, nil
	}
	if ownerID == userID {
		return 200, nil
	}
	return 409, nil
}

// Получает список заказов
func (s *Service) ListOrders(ctx context.Context, userID int64) ([]Order, error) {
	return s.repo.ListOrders(ctx, userID)
}
