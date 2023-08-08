package storage

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
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
	SetMetric(string, string, string) error
	GetValueByName(string, string) (string, error)
	GetAllData() string
}

type SysConfig struct {
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	ServerAddress  string `env:"ADDRESS"`
}

type Metrics struct {
	Alloc         Gauge
	BuckHashSys   Gauge
	Frees         Gauge
	GCCPUFraction Gauge
	GCSys         Gauge
	HeapAlloc     Gauge
	HeapIdle      Gauge
	HeapInuse     Gauge
	HeapObjects   Gauge
	HeapReleased  Gauge
	HeapSys       Gauge
	LastGC        Gauge
	Lookups       Gauge
	MCacheInuse   Gauge
	MCacheSys     Gauge
	MSpanInuse    Gauge
	MSpanSys      Gauge
	Mallocs       Gauge
	NextGC        Gauge
	NumForcedGC   Gauge
	NumGC         Gauge
	OtherSys      Gauge
	PauseTotalNs  Gauge
	StackInuse    Gauge
	StackSys      Gauge
	Sys           Gauge
	TotalAlloc    Gauge
	PollCount     Counter
	RandomValue   Gauge
}

func (m *Metrics) RenewMetrics() {
	var memstat runtime.MemStats
	runtime.ReadMemStats(&memstat)
	m.Alloc = Gauge(memstat.Alloc)
	m.BuckHashSys = Gauge(memstat.BuckHashSys)
	m.Frees = Gauge(memstat.Frees)
	m.GCCPUFraction = Gauge(memstat.GCCPUFraction)
	m.GCSys = Gauge(memstat.GCSys)
	m.HeapAlloc = Gauge(memstat.HeapAlloc)
	m.HeapIdle = Gauge(memstat.HeapIdle)
	m.HeapInuse = Gauge(memstat.HeapInuse)
	m.HeapObjects = Gauge(memstat.HeapObjects)
	m.HeapReleased = Gauge(memstat.HeapReleased)
	m.HeapSys = Gauge(memstat.HeapSys)
	m.LastGC = Gauge(memstat.LastGC)
	m.Lookups = Gauge(memstat.Lookups)
	m.MCacheInuse = Gauge(memstat.MCacheInuse)
	m.MCacheSys = Gauge(memstat.MCacheSys)
	m.MSpanInuse = Gauge(memstat.MSpanInuse)
	m.MSpanSys = Gauge(memstat.MSpanSys)
	m.Mallocs = Gauge(memstat.Mallocs)
	m.NextGC = Gauge(memstat.NextGC)
	m.NumForcedGC = Gauge(memstat.NumForcedGC)
	m.NumGC = Gauge(memstat.NumGC)
	m.OtherSys = Gauge(memstat.OtherSys)
	m.PauseTotalNs = Gauge(memstat.PauseTotalNs)
	m.StackInuse = Gauge(memstat.StackInuse)
	m.Sys = Gauge(memstat.Sys)
	m.TotalAlloc = Gauge(memstat.TotalAlloc)
	m.PollCount += 1
	m.RandomValue = Gauge(rand.Float64())
}

func (m *Metrics) SendMetrics(hostAndPort string) error {
	fieldsValues := reflect.ValueOf(m).Elem()
	fieldsTypes := reflect.TypeOf(m).Elem()
	for i := 0; i < fieldsValues.NumField()-2; i++ {
		fieldType := strings.Replace(fieldsTypes.Field(i).Type.String(), "storage.", "", -1)
		r, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/update/%s/%s/%v",
				hostAndPort,
				fieldType,
				fieldsTypes.Field(i).Name,
				fieldsValues.Field(i)),
			nil)
		if err != nil {
			return fmt.Errorf("there's an error in creating send metric request: type - %s, name - %s, value - %v, error - %e",
				fieldType,
				fieldsTypes.Field(i).Name,
				fieldsValues.Field(i),
				err,
			)

		}
		r.Header.Add("Content-Type", "text/plain")
		client := &http.Client{}
		res, err := client.Do(r)
		if err != nil {
			return fmt.Errorf("there's an error in sending request: %e", err)
		}
		err = res.Body.Close()
		if err != nil {
			return fmt.Errorf("error in closing res body - %e", err)
		}
	}
	return nil
}

func (m *MemStorage) SetMetric(metricType string, metricName string, metricValue string) error {
	if strings.ToUpper(metricType) == "GAUGE" {
		flt, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("can't convert to float64 %e", err)
		}
		m.PollCount += 1
		m.Gauge[metricName] = Gauge(flt)
		return nil
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
