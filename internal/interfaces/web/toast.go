package web

import "github.com/madalinpopa/gocost-web/internal/interfaces/web/respond"

type ToastType = respond.ToastType
type ToastMessage = respond.ToastMessage

const (
	Success  = respond.Success
	ErrorMsg = respond.ErrorMsg
	Warning  = respond.Warning
	Info     = respond.Info
)

func ToastEvent(t ToastType, message string) map[string]any {
	return respond.ToastEvent(t, message)
}
