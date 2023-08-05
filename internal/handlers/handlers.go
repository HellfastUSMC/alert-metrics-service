package handlers

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ReturnMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	url := strings.Split(req.URL.String(), "/")
	if len(url) < 4 {
		http.Error(res, "Bad url", http.StatusNotFound)
		return
	}
	updateUrl := storage.UpdateParse{}
	updateUrl.MetricType, updateUrl.MetricName = chi.URLParam(req, "metricType"), chi.URLParam(req, "metricName")
	if strings.ToUpper(updateUrl.MetricType) != "GAUGE" && strings.ToUpper(updateUrl.MetricType) != "COUNTER" {
		http.Error(res, "Wrong metric type", http.StatusBadRequest)
		return
	}
	val, err := storage.Store.GetValueByName(updateUrl.MetricType, updateUrl.MetricName)
	if err != nil {
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusNotFound)
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.Write([]byte(val))
}

func GetMetrics(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	url := strings.Split(req.URL.String(), "/")
	if len(url) < 5 {
		http.Error(res, "Bad url", http.StatusNotFound)
		return
	}
	updateUrl := storage.UpdateParse{}
	updateUrl.MetricType, updateUrl.MetricName, updateUrl.MetricVal = chi.URLParam(req, "metricType"), chi.URLParam(req, "metricName"), chi.URLParam(req, "metricValue")
	if strings.ToUpper(updateUrl.MetricType) != "GAUGE" && strings.ToUpper(updateUrl.MetricType) != "COUNTER" || updateUrl.MetricVal == "" {
		http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
		return
	}
	if strings.ToUpper(updateUrl.MetricType) == "GAUGE" {
		if _, err := strconv.ParseFloat(updateUrl.MetricVal, 64); err != nil {
			http.Error(res, "Can't parse metric value", http.StatusBadRequest)
			return
		}
	}
	if strings.ToUpper(updateUrl.MetricType) == "COUNTER" {
		if _, err := strconv.ParseInt(updateUrl.MetricVal, 10, 64); err != nil {
			http.Error(res, "Can't parse metric value", http.StatusBadRequest)
			return
		}
	}
	if err := storage.Store.SetMetric(updateUrl.MetricType, updateUrl.MetricName, updateUrl.MetricVal); err != nil {
		http.Error(res, fmt.Sprintf("Error occurred when converting to float64 or int64 - %e", err), http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "text/plain; charset=utf-8")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(200)
}

func GetAllStats(res http.ResponseWriter, req *http.Request) {
	allStats := []string{}
	for key, val := range storage.Store.Gauge {
		allStats = append(allStats, fmt.Sprintf("%s: %s", key, fmt.Sprintf("%f", val)))
	}
	for key, val := range storage.Store.Counter {
		allStats = append(allStats, fmt.Sprintf("%s: %s", key, fmt.Sprintf("%d", val)))
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.Write([]byte(strings.Join(allStats, "\n")))
}
