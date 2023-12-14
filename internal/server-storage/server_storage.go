// Package serverstorage Пакет для работы с серверным хранилищем метрик, содержит в себе описание структур и
// методов работы с ним
package serverstorage

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

// Gauge Тип данных для метрик
type Gauge float64

// Counter Тип данных для метрик
type Counter int64

// Dumper Интерфейс дампера для описания хранилищ
type Dumper interface {
	WriteDump([]byte) error
	ReadDump() ([]string, error)
	Ping() error
}

// MemStorage Структура хранилища метрик для сервера
type MemStorage struct {
	Dumper    Dumper         `json:"-"`
	Logger    logger.CLogger `json:"-"`
	Mutex     *sync.Mutex    `json:"-"`
	Gauge     map[string]Gauge
	Counter   map[string]Counter
	PollCount Counter
}

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

// MemStorekeeper Интерфейс для взаимодействия с хранилищем метрик
type MemStorekeeper interface {
	SetMetric(metricType string, metricName string, metricValue interface{}) error
	GetValueByName(metricType string, metricName string) (string, error)
	GetAllData() string
	Ping() error
}

// UpdateParse Структура для парса метрик из запроса
type UpdateParse struct {
	MetricType string
	MetricName string
	MetricVal  string
}

// ReadDump Функция обертка для чтения метрик из дампа
func (m *MemStorage) ReadDump() error {
	strs, err := m.Dumper.ReadDump()
	if err != nil {
		return err
	}
	m.Mutex.Lock()
	err = json.Unmarshal([]byte(strs[len(strs)-2]), m)
	m.Mutex.Unlock()
	if err != nil {
		return fmt.Errorf("can't unmarshal dump file - %e", err)
	}
	m.Logger.Info().Msg("Metrics received")
	return nil
}

// WriteDump Функция обертка для записи метрик в дамп
func (m *MemStorage) WriteDump() error {
	m.Mutex.Lock()
	jsonMemStore, err := json.Marshal(m)
	m.Mutex.Unlock()
	if err != nil {
		return fmt.Errorf("can't marshal dump data - %e", err)
	}
	err = m.Dumper.WriteDump(jsonMemStore)
	if err != nil {
		return fmt.Errorf("can't write dump data - %v", err)
	}
	return nil
}

// Ping Функция обертка для проверки соединения с БД
func (m *MemStorage) Ping() error {
	if m.Dumper != nil {
		if err := m.Dumper.Ping(); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("there's no dumpers avaliable")

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

func NewMemStorage(dumper Dumper, logger logger.CLogger) *MemStorage {
	return &MemStorage{
		Gauge:     map[string]Gauge{},
		Counter:   map[string]Counter{},
		PollCount: 0,
		Dumper:    dumper,
		Logger:    logger,
		Mutex:     &sync.Mutex{},
	}
}
