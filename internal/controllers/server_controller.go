package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"

	"github.com/go-chi/chi/v5"
)

type ServerController struct {
	Logger   CLogger
	Config   *config.SysConfig
	MemStore storage.MemStorekeeper
	CRouter  *chi.Mux
}

func (c *ServerController) returnMetric() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		url := strings.Split(req.URL.String(), "/")
		if len(url) < 4 {
			http.Error(res, "Bad url", http.StatusNotFound)
			return
		}
		updateURL := storage.UpdateParse{}
		updateURL.MetricType, updateURL.MetricName = chi.URLParam(req, "metricType"), chi.URLParam(req, "metricName")
		if strings.ToUpper(updateURL.MetricType) != "GAUGE" && strings.ToUpper(updateURL.MetricType) != "COUNTER" {
			http.Error(res, "Wrong metric type", http.StatusBadRequest)
			return
		}
		val, err := c.MemStore.GetValueByName(updateURL.MetricType, updateURL.MetricName)
		if err != nil {
			c.Logger.Errorf("error of GetValueByName ", err)
			http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusNotFound)
		}
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err = res.Write([]byte(val)); err != nil {
			c.Logger.Error(err)
		}
	}
}

func (c *ServerController) getMetrics() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		url := strings.Split(req.URL.String(), "/")
		if len(url) < 5 {
			http.Error(res, "Bad url", http.StatusNotFound)
			return
		}
		updateURL := storage.UpdateParse{}
		updateURL.MetricType, updateURL.MetricName, updateURL.MetricVal = chi.URLParam(req, "metricType"), chi.URLParam(req, "metricName"), chi.URLParam(req, "metricValue")
		if strings.ToUpper(updateURL.MetricType) != "GAUGE" && strings.ToUpper(updateURL.MetricType) != "COUNTER" || updateURL.MetricVal == "" {
			http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
			return
		}
		if strings.ToUpper(updateURL.MetricType) == "GAUGE" {
			if _, err := strconv.ParseFloat(updateURL.MetricVal, 64); err != nil {
				http.Error(res, "Can't parse metric value", http.StatusBadRequest)
				return
			}
		}
		if strings.ToUpper(updateURL.MetricType) == "COUNTER" {
			if _, err := strconv.ParseInt(updateURL.MetricVal, 10, 64); err != nil {
				http.Error(res, "Can't parse metric value", http.StatusBadRequest)
				return
			}
		}
		if err := c.MemStore.SetMetric(updateURL.MetricType, updateURL.MetricName, updateURL.MetricVal); err != nil {
			c.Logger.Errorf("error of SetMetric ", err)
			http.Error(res, fmt.Sprintf("Error occurred when converting to float64 or int64 - %e", err), http.StatusInternalServerError)
			return
		}
		res.Header().Add("content-type", "text/plain; charset=utf-8")
		res.Header().Add("Date", time.Now().Format(http.TimeFormat))
		res.WriteHeader(200)
	}

}

func (c *ServerController) getAllStats() func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, _ *http.Request) {
		allStats := c.MemStore.GetAllData()
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := res.Write([]byte(allStats)); err != nil {
			c.Logger.Error(err)
		}

	}
}

func (c *ServerController) Router() *chi.Mux {
	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Get("/", c.getAllStats())
		router.Route("/value", func(router chi.Router) {
			router.Get("/{metricType}/{metricName}", c.returnMetric())
		})
		router.Route("/update", func(router chi.Router) {
			router.Post("/{metricType}/{metricName}/{metricValue}", c.getMetrics())
		})
	})
	return router
}
