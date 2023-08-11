package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
)

type serverController struct {
	Logger   CLogger
	Config   *config.SysConfig
	MemStore serverstorage.MemStorekeeper
}

func (c *serverController) returnMetric(res http.ResponseWriter, req *http.Request) {
	updateURL := serverstorage.UpdateParse{}
	updateURL.MetricType = chi.URLParam(req, "metricType")
	updateURL.MetricName = chi.URLParam(req, "metricName")

	if strings.ToUpper(updateURL.MetricType) != "GAUGE" && strings.ToUpper(updateURL.MetricType) != "COUNTER" {
		http.Error(res, "Wrong metric type", http.StatusBadRequest)
		return
	}
	val, err := c.MemStore.GetValueByName(updateURL.MetricType, updateURL.MetricName)
	if err != nil {
		c.Errorf("error of GetValueByName ", err)
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusNotFound)
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	if _, err = res.Write([]byte(val)); err != nil {
		c.Error(err)
	}
}

func (c *serverController) getMetrics(res http.ResponseWriter, req *http.Request) {
	updateURL := serverstorage.UpdateParse{}
	updateURL.MetricType = chi.URLParam(req, "metricType")
	updateURL.MetricName = chi.URLParam(req, "metricName")
	updateURL.MetricVal = chi.URLParam(req, "metricValue")

	if strings.ToUpper(updateURL.MetricType) != "GAUGE" &&
		strings.ToUpper(updateURL.MetricType) != "COUNTER" ||
		updateURL.MetricVal == "" {
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
		http.Error(
			res,
			fmt.Sprintf("Error occurred when converting to float64 or int64 - %e", err),
			http.StatusInternalServerError,
		)
		return
	}
	res.Header().Add("content-type", "text/plain; charset=utf-8")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(200)
}

func (c *serverController) getAllStats(res http.ResponseWriter, _ *http.Request) {
	allStats := c.MemStore.GetAllData()
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	if _, err := res.Write([]byte(allStats)); err != nil {
		c.Error(err)
	}

}

func (c *serverController) Route() *chi.Mux {
	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Get("/", c.getAllStats)
		router.Get("/value/{metricType}/{metricName}", c.returnMetric)
		router.Post("/update/{metricType}/{metricName}/{metricValue}", c.getMetrics)
	})
	return router
}

func (c *serverController) Info(i interface{}) {
	c.Logger.Info(i)
}

func (c *serverController) Warn(i interface{}) {
	c.Logger.Warn(i)
}

func (c *serverController) Warning(i interface{}) {
	c.Logger.Warning(i)
}

func (c *serverController) Error(i interface{}) {
	c.Logger.Error(i)
}

func (c *serverController) Infof(s string, args ...interface{}) {
	c.Logger.Infof(s, args)
}

func (c *serverController) Warnf(s string, args ...interface{}) {
	c.Logger.Warnf(s, args)
}

func (c *serverController) Errorf(s string, args ...interface{}) {
	c.Logger.Errorf(s, args)
}

func NewServerController(logger CLogger, conf *config.SysConfig, mStore *serverstorage.MemStorage) *serverController {
	return &serverController{
		Logger:   logger,
		Config:   conf,
		MemStore: mStore,
	}
}
