package service

import (
	"context"
	"errors"
	"subscription_service/internal/domain"
	"testing"
)

// ---- repo mock ----

type repoMock struct {
	createFn    func(ctx context.Context, s domain.Subscription) (int64, error)
	getByIDFn   func(ctx context.Context, id int64) (*domain.Subscription, error)
	updateFn    func(ctx context.Context, s domain.Subscription) (*domain.Subscription, error)
	deleteFn    func(ctx context.Context, id int64) (bool, error)
	listFn      func(ctx context.Context, f domain.ListFilter) ([]domain.Subscription, error)
	totalCostFn func(ctx context.Context, f domain.TotalFilter) (int64, error)
}

func (m *repoMock) Create(ctx context.Context, s domain.Subscription) (int64, error) {
	if m.createFn == nil {
		panic("createFn is nil")
	}
	return m.createFn(ctx, s)
}

func (m *repoMock) GetByID(ctx context.Context, id int64) (*domain.Subscription, error) {
	if m.getByIDFn == nil {
		panic("getByIDFn is nil")
	}
	return m.getByIDFn(ctx, id)
}

func (m *repoMock) Update(ctx context.Context, s domain.Subscription) (*domain.Subscription, error) {
	if m.updateFn == nil {
		panic("updateFn is nil")
	}
	return m.updateFn(ctx, s)
}

func (m *repoMock) Delete(ctx context.Context, id int64) (bool, error) {
	if m.deleteFn == nil {
		panic("deleteFn is nil")
	}
	return m.deleteFn(ctx, id)
}

func (m *repoMock) List(ctx context.Context, f domain.ListFilter) ([]domain.Subscription, error) {
	if m.listFn == nil {
		panic("listFn is nil")
	}
	return m.listFn(ctx, f)
}

func (m *repoMock) TotalCost(ctx context.Context, f domain.TotalFilter) (int64, error) {
	if m.totalCostFn == nil {
		panic("totalCostFn is nil")
	}
	return m.totalCostFn(ctx, f)
}

var _ SubscriptionRepository = (*repoMock)(nil)

// ---- tests ----

func TestCreate_InvalidDateFormat_ReturnsErrInvalidInput(t *testing.T) {
	repo := &repoMock{
		createFn: func(ctx context.Context, s domain.Subscription) (int64, error) {
			t.Fatal("Create should not be called on invalid input")
			return 0, nil
		},
	}

	svc := NewSubscriptionService(repo)

	_, err := svc.Create(context.Background(), CreateSubscriptionRequest{
		ServiceName: "Netflix",
		Price:       1000,
		UserID:      "60610fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "2025-07", // WRONG, expected MM-YYYY
		EndDate:     nil,
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestDelete_NotFound_ReturnsErrNotFound(t *testing.T) {
	repo := &repoMock{
		deleteFn: func(ctx context.Context, id int64) (bool, error) {
			return false, nil // not deleted => not found
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.Delete(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_Deleted_OK(t *testing.T) {
	repo := &repoMock{
		deleteFn: func(ctx context.Context, id int64) (bool, error) {
			return true, nil
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestTotalCost_RepoError_Propagates(t *testing.T) {
	wantErr := errors.New("db down")

	repo := &repoMock{
		totalCostFn: func(ctx context.Context, f domain.TotalFilter) (int64, error) {
			return 0, wantErr
		},
	}

	svc := NewSubscriptionService(repo)

	from, _ := domain.ParseMonthYear("07-2025")
	to, _ := domain.ParseMonthYear("10-2025")

	_, err := svc.TotalCost(context.Background(), domain.TotalFilter{
		UserID:      nil,
		ServiceName: nil,
		From:        from,
		To:          to,
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
