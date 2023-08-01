package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type gauge float64
type counter int64

type MemStorage struct {
	Metrics   map[string]gauge
	PollCount counter
}

type updateParse struct {
	metricType string
	metricName string
	metricVal  string
}

var store = MemStorage{Metrics: map[string]gauge{}}

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
		http.Error(res, "Bad url", http.StatusNotFound)
		return
	}
	updateUrl := updateParse{}
	updateUrl.metricType = url[2]
	updateUrl.metricName = url[3]
	updateUrl.metricVal = url[4]
	fmt.Println(updateUrl)
	if strings.ToUpper(updateUrl.metricType) != "GAUGE" && strings.ToUpper(updateUrl.metricType) != "COUNTER" || updateUrl.metricVal == "" {
		http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
		return
	}
	if _, err := strconv.ParseFloat(updateUrl.metricVal, 64); err != nil {
		http.Error(res, "Wrong metric name", http.StatusBadRequest)
		return
	}
	if err := store.setMetric(updateUrl.metricName, updateUrl.metricVal); err != nil {
		http.Error(res, "Error occurred when converting to float64", http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "text/plain; charset=utf-8")
	res.Header().Add("Date", string(time.Now().Format(http.TimeFormat)))
	res.WriteHeader(200)
	fmt.Println(store)
}

func (m *MemStorage) setMetric(metricName string, metricValue string) error {
	flt, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return fmt.Errorf("can't convert to float64 %e", err)
	}
	m.PollCount += 1
	m.Metrics[metricName] = gauge(flt)
	return nil
}
