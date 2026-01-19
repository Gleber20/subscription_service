package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"subscription_service/internal/domain"
	"subscription_service/internal/service"

	"github.com/jmoiron/sqlx"
)

type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

var _ service.SubscriptionRepository = (*SubscriptionRepo)(nil)

// Create inserts subscription and returns new ID.
func (r *SubscriptionRepo) Create(ctx context.Context, s domain.Subscription) (int64, error) {
	var id int64
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id int64) (*domain.Subscription, error) {
	var s domain.Subscription
	err := r.db.GetContext(ctx, &s, `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// Update updates and returns the updated entity (re-read).
func (r *SubscriptionRepo) Update(ctx context.Context, s domain.Subscription) (*domain.Subscription, error) {
	_, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET service_name = $1,
		    price = $2,
		    user_id = $3,
		    start_date = $4,
		    end_date = $5,
		    updated_at = now()
		WHERE id = $6
	`, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate, s.ID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, s.ID)
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	return err
}

func (r *SubscriptionRepo) List(ctx context.Context, f domain.ListFilter) ([]domain.Subscription, error) {
	where, args := buildWhereList(f)

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		%s
		ORDER BY id
		LIMIT %d OFFSET %d
	`, where, limit, f.Offset)

	var items []domain.Subscription
	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

// TotalCost counts total monthly price for each month in [From..To] inclusive by month.
func (r *SubscriptionRepo) TotalCost(ctx context.Context, f domain.TotalFilter) (int64, error) {
	if err := f.Validate(); err != nil {
		return 0, err
	}

	// Convert inclusive "To" into exclusive upper bound
	toExclusive := f.ToExclusive()

	// optional filters for subscriptions alias "s"
	where, args := buildWhereBase(f.UserID, f.ServiceName, "s")

	// months series: from .. (toExclusive - 1 month)
	// Example: From=07-2025, To=10-2025 => toExclusive=11-2025 => series ends at 10-2025
	args = append(args, f.From, time.Date(toExclusive.Year(), toExclusive.Month()-1, 1, 0, 0, 0, 0, time.UTC))

	query := fmt.Sprintf(`
		WITH months AS (
			SELECT generate_series($%d::date, $%d::date, interval '1 month')::date AS m
		)
		SELECT COALESCE(SUM(s.price), 0) AS total
		FROM months
		JOIN subscriptions s
		  ON s.start_date <= months.m
		 AND (s.end_date IS NULL OR s.end_date >= months.m)
		%s
	`, len(args)-1, len(args), where)

	var total int64
	if err := r.db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

// ===== where builders =====

// Base filters (user_id, service_name). alias is "" or "s".
func buildWhereBase(userID, serviceName *string, alias string) (string, []any) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 2)

	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	if userID != nil && *userID != "" {
		args = append(args, *userID)
		clauses = append(clauses, fmt.Sprintf("%suser_id = $%d", prefix, len(args)))
	}
	if serviceName != nil && *serviceName != "" {
		args = append(args, *serviceName)
		clauses = append(clauses, fmt.Sprintf("%sservice_name = $%d", prefix, len(args)))
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

// List-specific where includes period intersection with inclusive To by month.
func buildWhereList(f domain.ListFilter) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)

	if f.UserID != nil && *f.UserID != "" {
		args = append(args, *f.UserID)
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if f.ServiceName != nil && *f.ServiceName != "" {
		args = append(args, *f.ServiceName)
		clauses = append(clauses, fmt.Sprintf("service_name = $%d", len(args)))
	}

	// Period intersection:
	// subscription intersects [From, ToExclusive)
	// Condition:
	//   (end_date IS NULL OR end_date >= From)
	//   AND start_date < ToExclusive
	if f.From != nil {
		from := domain.MonthStartUTC(*f.From)
		args = append(args, from)
		clauses = append(clauses, fmt.Sprintf("(end_date IS NULL OR end_date >= $%d)", len(args)))
	}
	if f.To != nil {
		toExclusive := domain.NextMonthStartUTC(*f.To)
		args = append(args, toExclusive)
		clauses = append(clauses, fmt.Sprintf("start_date < $%d", len(args)))
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}
