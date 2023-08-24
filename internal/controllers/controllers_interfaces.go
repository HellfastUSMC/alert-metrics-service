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

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
