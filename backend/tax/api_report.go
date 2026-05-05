package tax

import (
	"context"
	"fmt"
	"time"

	"encore.dev/beta/errs"
)

//encore:api public method=GET path=/companies/:id/reports/:yearMonth
func (s *Service) GetMonthReport(ctx context.Context, id, yearMonth string) (*MonthReport, error) {
	exists, err := companyExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &errs.Error{Code: errs.NotFound, Message: "company not found"}
	}

	start, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("yearMonth must be in YYYY-MM format, got %q", yearMonth),
		}
	}
	end := start.AddDate(0, 1, 0)

	return queryMonthReport(ctx, id, yearMonth, start, end)
}
