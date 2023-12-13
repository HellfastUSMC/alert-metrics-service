package controllers

type agentHandler interface {
	RenewMetrics()
	RenewMemCPUMetrics()
	SendBatchMetrics(key string, URL string, KeyPath string) error
}

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

// Metrics Структура метрики для преобразований JSON
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"` // Значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // Значение метрики в случае передачи gauge
	ID    string   `json:"id"`              // Имя метрики
	MType string   `json:"type"`            // Параметр, принимающий значение gauge или counter
}
