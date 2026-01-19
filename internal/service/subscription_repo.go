package service

import (
	"context"
	"subscription_service/internal/domain"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s domain.Subscription) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Subscription, error)
	Update(ctx context.Context, s domain.Subscription) (*domain.Subscription, error)
	Delete(ctx context.Context, id int64) error

	List(ctx context.Context, f domain.ListFilter) ([]domain.Subscription, error)
	TotalCost(ctx context.Context, f domain.TotalFilter) (int64, error)
}
