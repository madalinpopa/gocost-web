package form

import (
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// MailRX is a regular expression used to validate the format of email addresses.
var MailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

var usernameCharsRX = regexp.MustCompile(`^[a-zA-Z0-9_]*$`)

// NotBlank checks if the provided string is not empty or whitespace-only and returns true if valid, false otherwise.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars checks if the given string's length in runes is less than or equal to the specified maximum `n`.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedValue checks if the given value exists within the provided list of permitted values and returns a boolean result.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Equal checks if two comparable values are equal and returns true if they are, otherwise returns false.
func Equal[T comparable](value1, value2 T) bool {
	return value1 == value2
}

// ValidateDate checks if the provided date is non-zero and returns true if valid, otherwise returns false.
func ValidateDate(date time.Time) bool {
	return !date.IsZero()
}

// MinChars checks if the given string has at least the specified number of characters (n).
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Matches checks if the given string value matches the provided regular expression pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Number checks if the given string represents a valid integer number and returns true if valid, otherwise false.
func Number(value string) bool {
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

// UsernameCharsOnly checks if a string contains only allowed username characters.
func UsernameCharsOnly(value string) bool {
	return usernameCharsRX.MatchString(value)
}

// MaxNumber checks if the given value is less than or equal to the specified maximum limit.
func MaxNumber(value int, max int) bool {
	return value <= max
}

// ValidateRole checks if the given role string matches one of the permitted values: 'regular', 'admin', or 'specialist'.
func ValidateRole(role string) bool {
	return PermittedValue(role, "regular", "admin", "specialist")
}

// ValidateURL checks if the given string is a valid URL by attempting to parse it.
func ValidateURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// ValidateRating checks if the given rating is within the valid range (1-5).
func ValidateRating(rating int) bool {
	return rating >= 1 && rating <= 5
}

// ValidatePrice checks if the given price is a positive number.
func ValidatePrice(price float64) bool {
	return price >= 0
}

// ValidateDuration checks if the given duration is a positive integer.
func ValidateDuration(duration int) bool {
	return duration > 0
}

// ValidateCategoryType checks if the given category type is either 'event' or 'service'.
func ValidateCategoryType(categoryType string) bool {
	return PermittedValue(categoryType, "event", "service")
}

// ValidateEntityType checks if the given entity type is valid ('profile', 'service', 'event').
func ValidateEntityType(entityType string) bool {
	return PermittedValue(entityType, "profile", "service", "event")
}

// ValidateLinkType checks if the given link type is valid ('social', 'buy', 'external').
func ValidateLinkType(linkType string) bool {
	return PermittedValue(linkType, "social", "buy", "external")
}

// ValidateDateRange checks if the start date is before or equal to the end date.
func ValidateDateRange(startDate, endDate time.Time) bool {
	return !startDate.IsZero() && !endDate.IsZero() && (startDate.Before(endDate) || startDate.Equal(endDate))
}

// PositiveNumber checks if the given integer is positive.
func PositiveNumber(value int) bool {
	return value > 0
}

// PositiveFloat checks if the given float is positive.
func PositiveFloat(value float64) bool {
	return value > 0
}

// ValidDateString checks if the string matches the YYYY-MM-DD format.
func ValidDateString(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

// ValidMonthString checks if the string matches the YYYY-MM format.
func ValidMonthString(value string) bool {
	_, err := time.Parse("2006-01", value)
	return err == nil
}
