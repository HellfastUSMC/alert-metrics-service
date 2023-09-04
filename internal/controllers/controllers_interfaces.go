package controllers

type agentHandler interface {
	RenewMetrics()
	SendMetrics(URL string) error
	SendBatchMetrics(URL string) error
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
