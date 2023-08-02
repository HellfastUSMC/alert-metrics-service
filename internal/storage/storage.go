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
	Metrics   map[string]Gauge
	PollCount Counter
}

type UpdateParse struct {
	MetricType string
	MetricName string
	MetricVal  string
}

var Store = MemStorage{Metrics: map[string]Gauge{}}

type SysConfig struct {
	PollInterval   int64
	ReportInterval int64
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

func (m *MemStorage) SetMetric(metricName string, metricValue string) error {
	flt, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return fmt.Errorf("can't convert to float64 %e", err)
	}
	m.PollCount += 1
	m.Metrics[metricName] = Gauge(flt)
	return nil
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

func (m *Metrics) SendMetrics(urlAndPort string) {
	fieldsValues := reflect.ValueOf(m).Elem()
	fieldsTypes := reflect.TypeOf(m).Elem()
	for i := 0; i < fieldsValues.NumField()-2; i++ {
		fieldType := strings.Replace(fmt.Sprintf("%s", fieldsTypes.Field(i).Type), "storage.", "", -1)
		r, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/update/%s/%s/%v",
				urlAndPort,
				fieldType,
				fieldsTypes.Field(i).Name,
				fieldsValues.Field(i)),
			nil)
		if err != nil {
			fmt.Printf("there's an error in creating send metric request: type - %s, name - %s, value - %v, error - %e\n",
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
			fmt.Printf("there's an error in sending request: %e\n", err)
		}
		defer res.Body.Close()
	}
}
