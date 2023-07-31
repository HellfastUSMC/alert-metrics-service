package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

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
}
type updateParse struct {
	metricType string
	metricName string
	metricVal  string
}

var stor MemStorage

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", getGauge)
	err := http.ListenAndServe("localhost:8088", mux)
	if err != nil {
		fmt.Println(fmt.Errorf("there's an error in server starting - %e", err))
	}
}

func getGauge(res http.ResponseWriter, req *http.Request) {
	//fmt.Println(len(strings.Split(req.URL.String(), "/")))
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if req.Header.Get("Content-type") != "text/plain" {
		http.Error(res, "Wrong content type", http.StatusBadRequest)
		return
	}
	if len(strings.Split(req.URL.String(), "/")) < 5 {
		http.Error(res, "Wrong length URL", http.StatusNotFound)
		return
	}

	url := strings.Split(req.URL.String(), "/")
	var updateUrl updateParse
	updateUrl.metricType = url[2]
	updateUrl.metricName = url[3]
	updateUrl.metricVal = url[4]

	if updateUrl.metricType == "gauge" {
		val, _ := strconv.ParseFloat(url[len(url)-1], 64)
		stor.storeGauge = gauge(val)
		fmt.Println("Store gauge", stor.storeGauge)
	} else if updateUrl.metricType == "counter" {
		val, _ := strconv.ParseInt(url[len(url)-1], 10, 64)
		stor.storeCount = append(stor.storeCount, counter(val))
		fmt.Println("Store count", stor.storeCount)
	}
	//res.Write([]byte(fmt.Sprintf("%f", val)))
	res.WriteHeader(200)
}

//func getCounter(res http.ResponseWriter, req *http.Request) {
//	if req.Method != http.MethodPost || req.Header.Get("Content-type") != "text/plain" {
//		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//	//count =
//	//	res.Write([]byte(req.URL.String()))
//
//}
