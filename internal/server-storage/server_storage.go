package serverstorage

import (
	"fmt"
	"strconv"
	"strings"
)

type Gauge float64
type Counter int64

type MemStorage struct {
	Gauge     map[string]Gauge
	Counter   map[string]Counter
	PollCount Counter
}

type UpdateParse struct {
	MetricType string
	MetricName string
	MetricVal  string
}

type MemStorekeeper interface {
	SetMetric(metricType string, metricName string, metricValue string) error
	GetValueByName(string, string) (string, error)
	GetAllData() string
}

func (m *MemStorage) SetMetric(metricType string, metricName string, metricValue string) error {
	if strings.ToUpper(metricType) == "GAUGE" {
		flt, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't convert to float64 %e", err)
		}
		m.PollCount += 1
		if _, ok := m.Counter[metricName]; !ok {
			m.Gauge[metricName] = Gauge(flt)
			return nil
		} else {
			m.Gauge[metricName] += Gauge(flt)
			return nil
		}
	} else if strings.ToUpper(metricType) == "COUNTER" {
		integ, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("can't convert to int64 %e", err)
		}
		m.PollCount += 1
		if _, ok := m.Counter[metricName]; !ok {
			m.Counter[metricName] = Counter(integ)
			return nil
		} else {
			m.Counter[metricName] += Counter(integ)
			return nil
		}
	}
	return fmt.Errorf("metric with type %s not found", metricType)
}

func (m *MemStorage) GetValueByName(metricType string, metricName string) (string, error) {
	if strings.ToUpper(metricType) == "GAUGE" {
		if val, ok := m.Gauge[metricName]; !ok {
			return "", fmt.Errorf("there's no metric called %s", metricName)
		} else {
			return strconv.FormatFloat(float64(val), 'f', -1, 64), nil
		}
	} else if strings.ToUpper(metricType) == "COUNTER" {
		if val, ok := m.Counter[metricName]; !ok {
			return "", fmt.Errorf("there's no metric called %s", metricName)
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
		Gauge:   map[string]Gauge{},
		Counter: map[string]Counter{},
	}
}
