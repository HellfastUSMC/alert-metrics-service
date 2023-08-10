package agentstorage

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Gauge float64
type Counter int64

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

func NewMetricsStorage() *Metrics {
	return &Metrics{
		Alloc:         0,
		BuckHashSys:   0,
		Frees:         0,
		GCCPUFraction: 0,
		GCSys:         0,
		HeapAlloc:     0,
		HeapIdle:      0,
		HeapInuse:     0,
		HeapObjects:   0,
		HeapReleased:  0,
		HeapSys:       0,
		LastGC:        0,
		Lookups:       0,
		MCacheInuse:   0,
		MCacheSys:     0,
		MSpanInuse:    0,
		MSpanSys:      0,
		Mallocs:       0,
		NextGC:        0,
		NumForcedGC:   0,
		NumGC:         0,
		OtherSys:      0,
		PauseTotalNs:  0,
		StackInuse:    0,
		StackSys:      0,
		Sys:           0,
		TotalAlloc:    0,
		PollCount:     0,
		RandomValue:   0,
	}
}
