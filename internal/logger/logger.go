package logger

import "github.com/rs/zerolog"

// CLogger Интерефейс логгера для сервисов
type CLogger interface {
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
}
