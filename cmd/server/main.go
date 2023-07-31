package main

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type gauge float64
type counter int64

var metricNames = []string{
	"Alloc",
	"BuckHashSys",
	"Frees",
	"GCCPUFraction",
	"GCSys",
	"HeapAlloc",
	"HeapIdle",
	"HeapInuse",
	"HeapObjects",
	"HeapReleased",
	"HeapSys",
	"LastGC",
	"Lookups",
	"MCacheInuse",
	"MCacheSys",
	"MSpanInuse",
	"MSpanSys",
	"Mallocs",
	"NextGC",
	"NumForcedGC",
	"NumGC",
	"OtherSys",
	"PauseTotalNs",
	"StackInuse",
	"StackSys",
	"Sys",
	"TotalAlloc",
}

type MemStorage struct {
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
	RandValue     gauge
	PollCount     counter
}
type updateParse struct {
	metricType string
	metricName string
	metricVal  string
}

var stor MemStorage

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", getMetrics)
	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		fmt.Println(fmt.Errorf("there's an error in server starting - %e", err))
	}
}

func getMetrics(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	url := strings.Split(req.URL.String(), "/")
	if len(url) < 5 {
		http.Error(res, "Bad url", http.StatusBadRequest)
		return
	}
	var updateUrl updateParse
	updateUrl.metricType = url[2]
	updateUrl.metricName = url[3]
	updateUrl.metricVal = url[4]
	fmt.Println(updateUrl)
	if strings.ToUpper(updateUrl.metricType) != "GAUGE" && strings.ToUpper(updateUrl.metricType) != "COUNTER" || updateUrl.metricVal == "" {
		http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
		return
	}
	if req.Header.Get("Content-type") != "text/plain" {
		http.Error(res, "Content type should be text/plain", http.StatusBadRequest)
		return
	}
	if !checkMetricName(updateUrl.metricName, metricNames) {
		http.Error(res, "Wrong metric name", http.StatusBadRequest)
		return
	}
	if err := stor.setMetric(updateUrl.metricName, updateUrl.metricVal); err != nil {
		http.Error(res, "Error occurred when converting to float64", http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "text/plain; charset=utf-8")
	res.Header().Add("Date", string(time.Now().Format(http.TimeFormat)))
	res.WriteHeader(200)
	fmt.Println(stor)
}

func checkMetricName(metricName string, metricsList []string) bool {
	for _, val := range metricsList {
		if strings.ToUpper(metricName) == strings.ToUpper(val) {
			return true
		}
	}
	return false
}

func (m *MemStorage) setMetric(metricName string, metricValue string) error {
	for i := 0; i < len(metricNames); i++ {
		if strings.ToUpper(metricNames[i]) == strings.ToUpper(metricName) {
			field := reflect.ValueOf(m).Elem().FieldByName(metricNames[i])
			if field.CanSet() {
				flt, err := strconv.ParseFloat(metricValue, 64)
				if err != nil {
					return fmt.Errorf("can't convert to float64", err)
				}
				m.PollCount += 1
				field.SetFloat(flt)
			}
		}
	}
	return nil
}
