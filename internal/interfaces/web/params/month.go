package params

import (
	"net/http"
	"time"
)

// GetMonthParam extracts the month from the request query parameters.
// It returns the current date (defaulting to now if not provided),
// as well as the previous and next month dates.
func GetMonthParam(r *http.Request) (current, prev, next time.Time) {
	monthParam := r.URL.Query().Get("month")
	current, err := time.Parse("2006-01", monthParam)
	if err != nil {
		current = time.Now()
	}

	prev = current.AddDate(0, -1, 0)
	next = current.AddDate(0, 1, 0)

	return current, prev, next
}
