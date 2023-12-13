package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"github.com/HellfastUSMC/alert-metrics-service/internal/middlewares"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
)

type serverController struct {
	Logger   logger.CLogger
	Config   *config.SysConfig
	MemStore serverstorage.MemStorekeeper
}

func (c *serverController) returnJSONMetric(res http.ResponseWriter, req *http.Request) {
	updateMetric := Metrics{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Can't read request body", http.StatusInternalServerError)
		return
	}
	if err1 := json.Unmarshal(body, &updateMetric); err1 != nil {
		http.Error(res, "Can't parse JSON", http.StatusInternalServerError)
		return
	}

	if strings.ToUpper(updateMetric.MType) != GaugeStr && strings.ToUpper(updateMetric.MType) != CounterStr {
		http.Error(res, "Wrong metric type", http.StatusBadRequest)
		return
	}
	val, err2 := c.MemStore.GetValueByName(updateMetric.MType, updateMetric.ID)
	if err2 != nil {
		c.Logger.Error().Err(err).Msg("error of GetValueByName ")
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusNotFound)
		return
	}
	if strings.ToUpper(updateMetric.MType) == GaugeStr {
		flVal, err3 := strconv.ParseFloat(val, 64)
		if err3 != nil {
			c.Logger.Error().Err(err)
			http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusInternalServerError)
			return
		}
		updateMetric.Value = &flVal
	} else if strings.ToUpper(updateMetric.MType) == CounterStr {
		intVal, err4 := strconv.ParseInt(val, 10, 64)
		if err4 != nil {
			c.Logger.Error().Err(err)
			http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusInternalServerError)
			return
		}
		updateMetric.Delta = &intVal
	}
	jsonData, err := json.Marshal(updateMetric)
	if err != nil {
		http.Error(res, "can't write JSON", http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
	if _, err = res.Write(jsonData); err != nil {
		c.Logger.Error().Err(err)
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusInternalServerError)
		return
	}
}

func (c *serverController) getJSONMetrics(res http.ResponseWriter, req *http.Request) {
	updateMetric := Metrics{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "can't read request body", http.StatusInternalServerError)
		return
	}
	if err5 := json.Unmarshal(body, &updateMetric); err5 != nil {
		http.Error(res, fmt.Sprintf("can't unmarshal JSON %v", err), http.StatusInternalServerError)
		return
	}
	if (strings.ToUpper(updateMetric.MType) != GaugeStr &&
		strings.ToUpper(updateMetric.MType) != CounterStr) ||
		(updateMetric.Value == nil &&
			updateMetric.Delta == nil) {
		http.Error(res, "Wrong or empty metric type value", http.StatusBadRequest)
		return
	}

	if strings.ToUpper(updateMetric.MType) == GaugeStr {
		err6 := c.MemStore.SetMetric(updateMetric.MType, updateMetric.ID, updateMetric.Value)
		if err6 != nil {
			c.Logger.Error().Err(err6).Msg("error of SetMetric")
			http.Error(
				res,
				fmt.Sprintf("Error occurred when setting metric - %e", err),
				http.StatusInternalServerError,
			)
			return
		}
	}

	if strings.ToUpper(updateMetric.MType) == CounterStr {
		err7 := c.MemStore.SetMetric(updateMetric.MType, updateMetric.ID, updateMetric.Delta)
		if err7 != nil {
			c.Logger.Error().Err(err7).Msg("error of SetMetric")
			http.Error(
				res,
				fmt.Sprintf("Error occurred when setting metric - %e", err),
				http.StatusInternalServerError,
			)
			return
		}
		newMetricVal, err8 := c.MemStore.GetValueByName(updateMetric.MType, updateMetric.ID)
		if err8 != nil {
			c.Logger.Error().Err(err8).Msg("can't get new metric value")
		}
		intVal, err9 := strconv.ParseInt(newMetricVal, 10, 64)
		if err9 != nil {
			c.Logger.Error().Err(err9).Msg("can't parse new int64 metric value")
		}
		updateMetric.Delta = &intVal
	}
	jsonData, err10 := json.Marshal(updateMetric)
	if err10 != nil {
		http.Error(res, "can't marshal JSON", http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
	if _, err = res.Write(jsonData); err != nil {
		c.Logger.Error().Err(err)
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusInternalServerError)
		return
	}

}

func (c *serverController) returnMetric(res http.ResponseWriter, req *http.Request) {
	updateURL := serverstorage.UpdateParse{}
	updateURL.MetricType = chi.URLParam(req, "metricType")
	updateURL.MetricName = chi.URLParam(req, "metricName")

	if strings.ToUpper(updateURL.MetricType) != GaugeStr && strings.ToUpper(updateURL.MetricType) != CounterStr {
		http.Error(res, "Wrong metric type", http.StatusBadRequest)
		return
	}
	val, err := c.MemStore.GetValueByName(updateURL.MetricType, updateURL.MetricName)
	if err != nil {
		c.Logger.Error().Err(err).Msg("error of GetValueByName ")
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusNotFound)
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
	if _, err = res.Write([]byte(val)); err != nil {
		c.Logger.Error().Err(err)
	}
}

func (c *serverController) getMetrics(res http.ResponseWriter, req *http.Request) {
	updateURL := serverstorage.UpdateParse{}
	updateURL.MetricType = chi.URLParam(req, "metricType")
	updateURL.MetricName = chi.URLParam(req, "metricName")
	updateURL.MetricVal = chi.URLParam(req, "metricValue")

	if strings.ToUpper(updateURL.MetricType) != GaugeStr &&
		strings.ToUpper(updateURL.MetricType) != CounterStr ||
		updateURL.MetricVal == "" {
		http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
		return
	}
	if strings.ToUpper(updateURL.MetricType) == GaugeStr {
		if _, err := strconv.ParseFloat(updateURL.MetricVal, 64); err != nil {
			http.Error(res, "Can't parse metric value", http.StatusBadRequest)
			return
		}
	}
	if strings.ToUpper(updateURL.MetricType) == CounterStr {
		if _, err := strconv.ParseInt(updateURL.MetricVal, 10, 64); err != nil {
			http.Error(res, "Can't parse metric value", http.StatusBadRequest)
			return
		}
	}
	if err := c.MemStore.SetMetric(updateURL.MetricType, updateURL.MetricName, updateURL.MetricVal); err != nil {
		c.Logger.Error().Err(err).Msg("error of SetMetric")
		http.Error(
			res,
			fmt.Sprintf("Error occurred when converting to float64 or int64 - %e", err),
			http.StatusInternalServerError,
		)
		return
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
}

// NewServerController Функция инициализации нового контролера сервера
func NewServerController(logger logger.CLogger, conf *config.SysConfig, mStore *serverstorage.MemStorage) *serverController {
	return &serverController{
		Logger:   logger,
		Config:   conf,
		MemStore: mStore,
	}
}

func (c *serverController) getAllStats(res http.ResponseWriter, _ *http.Request) {
	allStats := c.MemStore.GetAllData()
	res.Header().Add("Content-Type", "text/html")
	if _, err := res.Write([]byte(allStats)); err != nil {
		c.Logger.Error().Err(err)
	}
}

func (c *serverController) getJSONMetricsBatch(res http.ResponseWriter, req *http.Request) {
	var updateMetrics []Metrics
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "can't read request body", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &updateMetrics)
	if err != nil {
		http.Error(res, "can't unmarshal json", http.StatusInternalServerError)
		return
	}
	for _, metric := range updateMetrics {
		if strings.ToUpper(metric.MType) == GaugeStr {
			if err := c.MemStore.SetMetric(metric.MType, metric.ID, metric.Value); err != nil {
				http.Error(
					res,
					fmt.Sprintf("error occured when set metric - %v", err),
					http.StatusInternalServerError,
				)
			}
			return
		} else {
			if err := c.MemStore.SetMetric(metric.MType, metric.ID, metric.Delta); err != nil {
				http.Error(
					res,
					fmt.Sprintf("error occured when set metric - %v", err),
					http.StatusInternalServerError,
				)
				return
			}
		}
	}
}

func (c *serverController) pingDB(res http.ResponseWriter, _ *http.Request) {
	err := c.MemStore.Ping()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
	} else {
		res.WriteHeader(http.StatusOK)
	}
}

// Route Функция для создания роутера сервера
func (c *serverController) Route() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middlewares.CheckHash(c.Logger))
	router.Use(middlewares.Gzip(c.Logger))
	router.Use(middlewares.ReqResLogging(c.Logger))
	router.Use(middlewares.CheckCert(c.Logger, c.Config.CryptoCert))
	router.Route("/", func(router chi.Router) {
		router.Get("/", c.getAllStats)
		router.Get("/ping", c.pingDB)
		router.Post("/value/", c.returnJSONMetric)
		router.Post("/update/", c.getJSONMetrics)
		router.Post("/updates/", c.getJSONMetricsBatch)
		router.Get("/value/{metricType}/{metricName}", c.returnMetric)
		router.Post("/update/{metricType}/{metricName}/{metricValue}", c.getMetrics)
		router.Get("/debug/pprof/", pprof.Index)
		router.Get("/debug/pprof/cmdline", pprof.Cmdline)
		router.Get("/debug/pprof/profile", pprof.Profile)
		router.Get("/debug/pprof/symbol", pprof.Symbol)
		router.Get("/debug/pprof/trace", pprof.Trace)
	})
	return router
}
