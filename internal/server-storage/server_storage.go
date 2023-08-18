package serverstorage

import (
	"fmt"
	"reflect"
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

type MemStorekeeper interface {
	SetMetric(metricType string, metricName string, metricValue interface{}) error
	GetValueByName(string, string) (string, error)
	GetAllData() string
}

type UpdateParse struct {
	MetricType string
	MetricName string
	MetricVal  string
}

func (m *MemStorage) SetMetric(metricType string, metricName string, metricValue interface{}) error {

	if strings.ToUpper(metricType) == "GAUGE" {
		m.PollCount += 1
		if _, ok := m.Counter[metricName]; !ok {
			if reflect.TypeOf(metricValue).String() == "string" {
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
	} else if strings.ToUpper(metricType) == "COUNTER" {
		m.PollCount += 1
		if _, ok := m.Counter[metricName]; !ok {
			if reflect.TypeOf(metricValue).String() == "string" {
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
			if reflect.TypeOf(metricValue).String() == "string" {
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
