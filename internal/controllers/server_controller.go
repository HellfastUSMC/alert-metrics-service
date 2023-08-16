package controllers

import (
	"fmt"
	"github.com/rs/zerolog"
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
type logRespWriter struct {
	data struct {
		code int
		size int
	}
	wr http.ResponseWriter
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
		c.Error().Err(err).Msg("error of GetValueByName ")
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusNotFound)
	}
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	if _, err = res.Write([]byte(val)); err != nil {
		c.Error().Err(err)
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
		c.Logger.Error().Err(err).Msg("error of SetMetric")
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
		c.Error().Err(err)
	}

}

func (c *serverController) Route() *chi.Mux {
	router := chi.NewRouter()
	router.Use(c.reqResLogging)
	router.Route("/", func(router chi.Router) {
		router.Get("/", c.getAllStats)
		router.Get("/value/{metricType}/{metricName}", c.returnMetric)
		router.Post("/update/{metricType}/{metricName}/{metricValue}", c.getMetrics)
	})
	return router
}

func (c *serverController) Info() *zerolog.Event {
	return c.Logger.Info()
}

func (c *serverController) Warn() *zerolog.Event {
	return c.Logger.Warn()
}

func (c *serverController) Error() *zerolog.Event {
	return c.Logger.Error()
}

func NewServerController(logger CLogger, conf *config.SysConfig, mStore *serverstorage.MemStorage) *serverController {
	return &serverController{
		Logger:   logger,
		Config:   conf,
		MemStore: mStore,
	}
}

func (c *serverController) reqResLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		rw := logRespWriter{
			data: struct {
				code int
				size int
			}{code: 0, size: 0},
			wr: res,
		}

		h.ServeHTTP(&rw, r)

		duration := time.Since(start).String()

		c.Logger.Info().Str("URI", uri).Str("method", method).Str("duration", duration).Int("code", rw.data.code).Int("size", rw.data.size)
	})
}

func (r *logRespWriter) Write(b []byte) (int, error) {
	size, err := r.wr.Write(b)
	r.data.size += size
	return size, err
}

func (r *logRespWriter) WriteHeader(statusCode int) {
	r.wr.WriteHeader(statusCode)
	r.data.code = statusCode
}

func (r *logRespWriter) Header() http.Header {
	header := r.wr.Header()
	return header
}
