package controllers

type agentHandler interface {
	RenewMetrics()
	RenewMemCPUMetrics()
	SendBatchMetrics(key string, URL string) error
}

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

// Metrics Структура метрики для преобразований JSON
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
