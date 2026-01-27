package service

import (
	"context"
	"errors"
	"fmt"
	"subscription_service/internal/domain"
	"time"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidDateRange = errors.New("invalid date range")
)

type SubscriptionService struct {
	repo SubscriptionRepository
}

func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

type CreateSubscriptionRequest struct {
	ServiceName string
	Price       int64
	UserID      string
	StartDate   string  // "MM-YYYY"
	EndDate     *string // nil = не задана
}

// PATCH: end_date — 3 состояния: не прислали / прислали null / прислали значение
type EndDateUpdate struct {
	Provided bool
	Value    *string
}

func EndDateNotProvided() EndDateUpdate {
	return EndDateUpdate{Provided: false}
}

func EndDateSetNull() EndDateUpdate {
	return EndDateUpdate{Provided: true, Value: nil}
}

func EndDateSetValue(v string) EndDateUpdate {
	return EndDateUpdate{Provided: true, Value: &v}
}

type UpdateSubscriptionRequest struct {
	ServiceName *string
	Price       *int64
	UserID      *string
	StartDate   *string
	EndDate     EndDateUpdate
}

func (s *SubscriptionService) Create(ctx context.Context, req CreateSubscriptionRequest) (int64, error) {
	if req.ServiceName == "" || req.UserID == "" || req.Price < 0 || req.StartDate == "" {
		return 0, fmt.Errorf("%w: required fields missing", ErrInvalidInput)
	}

	start, err := parseMonthYear(req.StartDate)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid start_date", ErrInvalidInput)
	}

	var end *time.Time
	if req.EndDate != nil {
		e, err := parseMonthYear(*req.EndDate)
		if err != nil {
			return 0, fmt.Errorf("%w: invalid end_date", ErrInvalidInput)
		}
		if e.Before(start) {
			return 0, fmt.Errorf("%w: end_date before start_date", ErrInvalidInput)
		}
		end = &e
	}

	sub := domain.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   start,
		EndDate:     end,
	}

	return s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id int64) (*domain.Subscription, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Update(ctx context.Context, id int64, req UpdateSubscriptionRequest) (*domain.Subscription, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidInput)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	// применяем PATCH
	if req.ServiceName != nil {
		existing.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, fmt.Errorf("%w: price must be >= 0", ErrInvalidInput)
		}
		existing.Price = *req.Price
	}
	if req.UserID != nil {
		if *req.UserID == "" {
			return nil, fmt.Errorf("%w: user_id empty", ErrInvalidInput)
		}
		existing.UserID = *req.UserID
	}
	if req.StartDate != nil {
		start, err := parseMonthYear(*req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid start_date", ErrInvalidInput)
		}
		existing.StartDate = start
	}

	if req.EndDate.Provided {
		if req.EndDate.Value == nil {
			existing.EndDate = nil
		} else {
			end, err := parseMonthYear(*req.EndDate.Value)
			if err != nil {
				return nil, fmt.Errorf("%w: invalid end_date", ErrInvalidInput)
			}
			existing.EndDate = &end
		}
	}

	// финальная проверка диапазона дат
	if existing.EndDate != nil && existing.EndDate.Before(existing.StartDate) {
		return nil, fmt.Errorf("%w: end_date before start_date", ErrInvalidInput)
	}

	return s.repo.Update(ctx, *existing)
}

func (s *SubscriptionService) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	return nil
}

func (s *SubscriptionService) List(ctx context.Context, f domain.ListFilter) ([]domain.Subscription, error) {
	// тут можно добавить валидацию фильтров, но сейчас держим просто
	return s.repo.List(ctx, f)
}

func (s *SubscriptionService) TotalCost(ctx context.Context, f domain.TotalFilter) (int64, error) {
	return s.repo.TotalCost(ctx, f)
}

func parseMonthYear(v string) (time.Time, error) {
	// формат "MM-YYYY"
	t, err := time.Parse("01-2006", v)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
