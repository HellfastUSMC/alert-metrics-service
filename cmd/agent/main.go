package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type gauge float64
type counter int64

type sysConfig struct {
	pollInterval   int64
	reportInterval int64
}

type metrics struct {
	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge
	HeapAlloc     gauge
	HeapIdle      gauge
	HeapInuse     gauge
	HeapObjects   gauge
	HeapReleased  gauge
	HeapSys       gauge
	LastGC        gauge
	Lookups       gauge
	MCacheInuse   gauge
	MCacheSys     gauge
	MSpanInuse    gauge
	MSpanSys      gauge
	Mallocs       gauge
	NextGC        gauge
	NumForcedGC   gauge
	NumGC         gauge
	OtherSys      gauge
	PauseTotalNs  gauge
	StackInuse    gauge
	StackSys      gauge
	Sys           gauge
	TotalAlloc    gauge
	PollCount     counter
	RandomValue   gauge
}

func main() {
	var stats metrics
	conf := sysConfig{
		pollInterval:   2,
		reportInterval: 1,
	}
	for {
		time.Sleep(time.Duration(conf.pollInterval) * time.Second)
		stats.renewMetrics()
		time.Sleep(time.Duration(conf.reportInterval-conf.pollInterval) * time.Second)
		stats.sendMetrics()
	}
}

func (m *metrics) renewMetrics() {
	var memstat runtime.MemStats
	runtime.ReadMemStats(&memstat)
	m.Alloc = gauge(memstat.Alloc)
	m.BuckHashSys = gauge(memstat.BuckHashSys)
	m.Frees = gauge(memstat.Frees)
	m.GCCPUFraction = gauge(memstat.GCCPUFraction)
	m.GCSys = gauge(memstat.GCSys)
	m.HeapAlloc = gauge(memstat.HeapAlloc)
	m.HeapIdle = gauge(memstat.HeapIdle)
	m.HeapInuse = gauge(memstat.HeapInuse)
	m.HeapObjects = gauge(memstat.HeapObjects)
	m.HeapReleased = gauge(memstat.HeapReleased)
	m.HeapSys = gauge(memstat.HeapSys)
	m.LastGC = gauge(memstat.LastGC)
	m.Lookups = gauge(memstat.Lookups)
	m.MCacheInuse = gauge(memstat.MCacheInuse)
	m.MCacheSys = gauge(memstat.MCacheSys)
	m.MSpanInuse = gauge(memstat.MSpanInuse)
	m.MSpanSys = gauge(memstat.MSpanSys)
	m.Mallocs = gauge(memstat.Mallocs)
	m.NextGC = gauge(memstat.NextGC)
	m.NumForcedGC = gauge(memstat.NumForcedGC)
	m.NumGC = gauge(memstat.NumGC)
	m.OtherSys = gauge(memstat.OtherSys)
	m.PauseTotalNs = gauge(memstat.PauseTotalNs)
	m.StackInuse = gauge(memstat.StackInuse)
	m.Sys = gauge(memstat.Sys)
	m.TotalAlloc = gauge(memstat.TotalAlloc)
	m.PollCount += 1
	m.RandomValue = gauge(rand.Float64())
}

func (m *metrics) sendMetrics() {
	fieldsValues := reflect.ValueOf(m).Elem()
	fieldsTypes := reflect.TypeOf(m).Elem()
	//fmt.Println(fieldsValues)
	//fmt.Println(fieldsTypes)
	for i := 0; i < fieldsValues.NumField()-2; i++ {
		fieldType := strings.Replace(fmt.Sprintf("%s", fieldsTypes.Field(i).Type), "main.", "", -1)
		//fmt.Println(fieldsValues.Field(i), fieldsTypes.Field(i).Name, fieldsTypes.Field(i).Type)
		if _, err := http.Post(
			fmt.Sprintf("http://localhost:8080/update/%s/%s/%v",
				fieldType,
				fieldsTypes.Field(i).Name,
				fieldsValues.Field(i)),
			"text/plain",
			nil); err != nil {
			fmt.Printf("there's an error in sending metric: type - %s, name - %s, value - %s, error - %e\n",
				fieldType,
				fieldsTypes.Field(i).Name,
				fieldsValues.Field(i),
				err,
			)
		}
	}
}
