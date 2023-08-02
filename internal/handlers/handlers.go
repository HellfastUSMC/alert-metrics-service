package handlers

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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
	updateUrl.MetricType = url[2]
	updateUrl.MetricName = url[3]
	updateUrl.MetricVal = url[4]
	if strings.ToUpper(updateUrl.MetricType) != "GAUGE" && strings.ToUpper(updateUrl.MetricType) != "COUNTER" || updateUrl.MetricVal == "" {
		http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
		return
	}
	if _, err := strconv.ParseFloat(updateUrl.MetricVal, 64); err != nil {
		http.Error(res, "Wrong metric name", http.StatusBadRequest)
		return
	}
	if err := storage.Store.SetMetric(updateUrl.MetricName, updateUrl.MetricVal); err != nil {
		http.Error(res, "Error occurred when converting to float64", http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "text/plain; charset=utf-8")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(200)
	fmt.Println(storage.Store)
}
