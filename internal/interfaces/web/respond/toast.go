package respond

type ToastType string

const (
	Success  ToastType = "success"
	ErrorMsg ToastType = "error"
	Warning  ToastType = "warning"
	Info     ToastType = "info"
)

type ToastMessage struct {
	Type    ToastType
	Message string
	Style   string
	ID      string
}

func ToastEvent(t ToastType, message string) map[string]any {
	return map[string]any{
		"showToast": map[string]string{
			"level":   string(t),
			"message": message,
		},
	}
}
