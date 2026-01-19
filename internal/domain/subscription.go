package domain

import (
	"fmt"
	"time"
)

const MonthYearLayout = "01-2006" // MM-YYYY

func ParseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse(MonthYearLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format %q (expected MM-YYYY): %w", s, err)
	}
	return MonthStartUTC(t), nil
}

func MonthStartUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func NextMonthStartUTC(t time.Time) time.Time {
	t = MonthStartUTC(t)
	return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
}

func FormatMonthYear(t time.Time) string {
	return MonthStartUTC(t).Format(MonthYearLayout)
}

type Subscription struct {
	ID          int64      `db:"id" json:"id"`
	ServiceName string     `db:"service_name" json:"service_name"`
	Price       int64      `db:"price" json:"price"`     // rubles per month
	UserID      string     `db:"user_id" json:"user_id"` // UUID as string
	StartDate   time.Time  `db:"start_date" json:"-"`
	EndDate     *time.Time `db:"end_date" json:"-"` // month start or NULL
	CreatedAt   time.Time  `db:"created_at" json:"-"`
	UpdatedAt   time.Time  `db:"updated_at" json:"-"`
}

type SubscriptionDTO struct {
	ID          int64   `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int64   `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"` // MM-YYYY
	EndDate     *string `json:"end_date,omitempty"`
}

func ToDTO(s Subscription) SubscriptionDTO {
	var end *string
	if s.EndDate != nil {
		v := FormatMonthYear(*s.EndDate)
		end = &v
	}

	return SubscriptionDTO{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   FormatMonthYear(s.StartDate),
		EndDate:     end,
	}
}
