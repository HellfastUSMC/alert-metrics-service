package agentstorage

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
)

type Gauge float64
type Counter int64

type Metric struct {
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

const (
	gaugeStr   = "GAUGE"
	counterStr = "COUNTER"
)

func (m *Metric) RenewMetrics() {
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

func (m *Metric) SendBatchMetrics(hostAndPort string) error {
	fieldsValues := reflect.ValueOf(m).Elem()
	fieldsTypes := reflect.TypeOf(m).Elem()
	var metricsList []controllers.Metrics
	var fieldType string
	for i := 0; i < fieldsValues.NumField(); i++ {
		if strings.Contains(strings.ToUpper(fieldsTypes.Field(i).Type.String()), gaugeStr) {
			fieldType = strings.ToLower(gaugeStr)
		}
		if strings.Contains(strings.ToUpper(fieldsTypes.Field(i).Type.String()), counterStr) {
			fieldType = strings.ToLower(counterStr)
		}
		metricStruct := controllers.Metrics{ID: fieldsTypes.Field(i).Name, MType: fieldType}
		if strings.ToUpper(metricStruct.MType) == gaugeStr {
			flVal := fieldsValues.Field(i).Float()
			metricStruct.Value = &flVal
		} else {
			intVal := fieldsValues.Field(i).Int()
			metricStruct.Delta = &intVal
		}
		metricsList = append(metricsList, metricStruct)
	}
	if metricsList == nil {
		return fmt.Errorf("nothing to send, metrics list is empty")
	}
	jsonByte, _ := json.Marshal(metricsList)
	var buff bytes.Buffer
	w, err := flate.NewWriter(&buff, flate.BestCompression)
	if err != nil {
		return fmt.Errorf("can't create new writer - %w", err)
	}

	_, err = w.Write(jsonByte)
	if err != nil {
		return fmt.Errorf("can't write compress JSON in gzip - %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("can't close writer - %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/updates/", hostAndPort),
		&buff,
	)
	if err != nil {
		return fmt.Errorf("there's an error in creating send metric request: type - %s, error - %w",
			fieldType,
			err,
		)
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept-Encoding", "gzip")
	r.Header.Add("Content-Encoding", "gzip")

	client := &http.Client{Timeout: time.Second * 2}
	res, err := client.Do(r)
	if err != nil {
		return fmt.Errorf("there's an error in sending request: %w", err)
	}

	err = res.Body.Close()
	if err != nil {
		return fmt.Errorf("error in closing res body - %w", err)
	}
	return nil
}

func NewMetricsStorage() *Metric {
	return &Metric{
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

func checkErr(errorsToRetry []any, err error) bool {
	for _, cErr := range errorsToRetry {
		if errors.As(err, &cErr) {
			return true
		}
	}
	return false
}

func RetryFunc(logger logger.CLogger, intervals []int, errorsToRetry []any, function func() error) error {
	err := function()
	if err != nil && checkErr(errorsToRetry, err) {
		for i, interval := range intervals {
			logger.Info().Msg(fmt.Sprintf("Error %v. Attempt #%d with interval %ds", err, i, interval))
			time.Sleep(time.Second * time.Duration(interval))
			errOK := checkErr(errorsToRetry, err)
			if errOK {
				err = function()
				if err == nil {
					return nil
				}
			}
		}
	}
	return err
}
