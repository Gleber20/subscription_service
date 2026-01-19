package domain

import (
	"fmt"
	"time"
)

// ListFilter is used for listing subscriptions.
// From/To are month starts (UTC). If both are set, period is interpreted as:
// [From, NextMonthStart(To))  (i.e. To is inclusive by month)
type ListFilter struct {
	UserID      *string
	ServiceName *string

	From *time.Time // optional, month start
	To   *time.Time // optional, month start (inclusive by month)

	Limit  int
	Offset int
}

// TotalFilter is used for aggregation.
// From/To are REQUIRED month starts (UTC). To is inclusive by month.
// Internally, you should convert To to ToExclusive = NextMonthStartUTC(To).
type TotalFilter struct {
	UserID      *string
	ServiceName *string

	From time.Time // month start (UTC)
	To   time.Time // month start (UTC), inclusive by month
}

// Validate checks obvious mistakes early (service layer can call this).
func (f TotalFilter) Validate() error {
	from := MonthStartUTC(f.From)
	to := MonthStartUTC(f.To)

	if to.Before(from) {
		return fmt.Errorf("invalid date range: 'to' must be >= 'from'")
	}
	return nil
}

// ToExclusive converts inclusive-month To into exclusive upper bound.
// Example: From=07-2025, To=10-2025 -> ToExclusive=11-2025
func (f TotalFilter) ToExclusive() time.Time {
	return NextMonthStartUTC(f.To)
}
