package tracking

import "errors"

var (
	ErrEmptyName          = errors.New("name cannot be empty")
	ErrNameTooLong        = errors.New("name exceeds maximum length of 100 characters")
	ErrDescriptionTooLong = errors.New("description exceeds maximum length")
	ErrInvalidMonth       = errors.New("month must be in YYYY-MM format")
	ErrEndMonthBeforeStartMonth = errors.New("end month must be after or equal to start month")
	ErrEndMonthNotAllowed = errors.New("end month is only allowed for recurrent categories")
	ErrCategoryNameExists = errors.New("category name already exists in group")
	ErrCategoryGroupMismatch = errors.New("category does not belong to this group")
	ErrGroupNotFound      = errors.New("group not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrInvalidOrder       = errors.New("order cannot be negative")
)
