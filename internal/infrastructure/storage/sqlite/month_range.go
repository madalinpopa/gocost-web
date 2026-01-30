package sqlite

import (
	"fmt"
	"time"
)

func monthToDateRange(month string) (start, end time.Time, err error) {
	start, err = time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month format %q (expected YYYY-MM): %w", month, err)
	}
	end = start.AddDate(0, 1, 0)
	return
}
