package form

import (
	"net/http"
	"slices"

	"github.com/go-playground/form/v4"
)

type Former interface {
	IsValid() bool
	AddFieldError(key, message string)
	AddNonFieldError(message string)
	CheckField(ok bool, key, message string)
	Validate()
}

type Base struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

func (b *Base) IsValid() bool {
	return len(b.FieldErrors) == 0 && len(b.NonFieldErrors) == 0
}

func (b *Base) AddFieldError(key, message string) {
	if b.FieldErrors == nil {
		b.FieldErrors = make(map[string]string)
	}
	b.FieldErrors[key] = message
}

func (b *Base) AddNonFieldError(message string) {
	if !slices.Contains(b.NonFieldErrors, message) {
		b.NonFieldErrors = append(b.NonFieldErrors, message)
	}
}

func (b *Base) CheckField(ok bool, key, message string) {
	if !ok {
		b.AddFieldError(key, message)
	}
}

func ParseAndValidateForm(r *http.Request, d *form.Decoder, f Former) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := d.Decode(f, r.PostForm); err != nil {
		return err
	}

	f.Validate()
	return nil
}
