package serverstorage

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

type Gauge float64
type Counter int64

type Dumper interface {
	WriteDump() error
	ReadDump() error
}

type MemStorage struct {
	Gauge     map[string]Gauge
	Counter   map[string]Counter
	PollCount Counter
	Dumper    Dumper `json:"-"`
}

type CLogger interface {
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
}

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

type MemStorekeeper interface {
	SetMetric(metricType string, metricName string, metricValue interface{}) error
	GetValueByName(metricType string, metricName string) (string, error)
	GetAllData() string
}

type UpdateParse struct {
	MetricType string
	MetricName string
	MetricVal  string
}

func (m *MemStorage) SetMetric(metricType string, metricName string, metricValue interface{}) error {
	if strings.ToUpper(metricType) == GaugeStr {
		m.PollCount += 1
		if fmt.Sprintf("%T", metricValue) == "string" {
			flt, err := strconv.ParseFloat(metricValue.(string), 64)
			if err != nil {
				return fmt.Errorf("can't convert to float64 %e", err)
			}
			m.Gauge[metricName] = Gauge(flt)
			return nil
		}
		m.Gauge[metricName] = Gauge(reflect.ValueOf(metricValue).Elem().Float())
		return nil
	}
	if strings.ToUpper(metricType) == CounterStr {
		m.PollCount += 1
		if _, ok := m.Counter[metricName]; !ok {
			if fmt.Sprintf("%T", metricValue) == "string" {
				integ, err := strconv.ParseInt(metricValue.(string), 10, 64)
				if err != nil {
					return fmt.Errorf("can't convert to int64 %e", err)
				}
				m.Counter[metricName] = Counter(integ)
				return nil
			}
			m.Counter[metricName] = Counter(reflect.ValueOf(metricValue).Elem().Int())
			return nil
		} else {
			if fmt.Sprintf("%T", metricValue) == "string" {
				integ, err := strconv.ParseInt(metricValue.(string), 10, 64)
				if err != nil {
					return fmt.Errorf("can't convert to int64 %e", err)
				}
				m.Counter[metricName] += Counter(integ)
				return nil
			}
			m.Counter[metricName] += Counter(reflect.ValueOf(metricValue).Elem().Int())
			return nil
		}
	}
	return fmt.Errorf("metric with type %s not found", metricType)
}

func (m *MemStorage) GetValueByName(metricType string, metricName string) (string, error) {
	if strings.ToUpper(metricType) == GaugeStr {
		if val, ok := m.Gauge[metricName]; !ok {
			return "", fmt.Errorf("there's no gauge metric called %s", metricName)
		} else {
			return strconv.FormatFloat(float64(val), 'f', -1, 64), nil
		}
	} else if strings.ToUpper(metricType) == CounterStr {
		if val, ok := m.Counter[metricName]; !ok {
			return "", fmt.Errorf("there's no counter metric called %s", metricName)
		} else {
			return strconv.FormatInt(int64(val), 10), nil
		}
	}
	return "", fmt.Errorf("metric with type %s not found", metricType)
}

func (m *MemStorage) GetAllData() string {
	allStats := []string{}
	for key, val := range m.Gauge {
		allStats = append(allStats, fmt.Sprintf("%s: %s", key, fmt.Sprintf("%f", val)))
	}
	for key, val := range m.Counter {
		allStats = append(allStats, fmt.Sprintf("%s: %s", key, fmt.Sprintf("%d", val)))
	}
	return strings.Join(allStats, "\n")
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:     map[string]Gauge{},
		Counter:   map[string]Counter{},
		PollCount: 0,
	}
}
