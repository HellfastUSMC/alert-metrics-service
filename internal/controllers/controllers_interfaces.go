package controllers

import "github.com/rs/zerolog"

type CLogger interface {
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
}

type agentHandler interface {
	RenewMetrics()
	SendMetrics(string) error
}

//
//type logTyper interface {
//	Str(key string, val string) *logTyper
//}
