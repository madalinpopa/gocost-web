package sqlite

import "time"

func monthToDateRange(month string) (start, end time.Time, err error) {
	start, err = time.Parse("2006-01", month)
	if err != nil {
		return
	}
	end = start.AddDate(0, 1, 0)
	return
}
