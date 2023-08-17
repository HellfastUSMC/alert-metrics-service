package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
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

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (c *serverController) returnMetric(res http.ResponseWriter, req *http.Request) {
	updateMetric := Metrics{}
	body, err := io.ReadAll(req.Body)
	if err := json.Unmarshal(body, &updateMetric); err != nil {
		http.Error(res, "Can't parse JSON", http.StatusInternalServerError)
		return
	}
	fmt.Println(updateMetric)

	if strings.ToUpper(updateMetric.MType) != "GAUGE" && strings.ToUpper(updateMetric.MType) != "COUNTER" {
		http.Error(res, "Wrong metric type", http.StatusBadRequest)
		return
	}
	val, err := c.MemStore.GetValueByName(updateMetric.MType, updateMetric.ID)
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
	updateMetric := Metrics{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "can't read request body", http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &updateMetric); err != nil {
		http.Error(res, "can't parse JSON", http.StatusInternalServerError)
		return
	}
	if strings.ToUpper(updateMetric.MType) == "GAUGE" &&
		updateMetric.Value == nil ||
		strings.ToUpper(updateMetric.MType) == "COUNTER" &&
			updateMetric.Delta == nil {
		http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
		return
	}

	if strings.ToUpper(updateMetric.MType) == "GAUGE" {
		err := c.MemStore.SetMetric(updateMetric.MType, updateMetric.ID, updateMetric.Value)
		if err != nil {
			c.Logger.Error().Err(err).Msg("error of SetMetric")
			http.Error(
				res,
				fmt.Sprintf("Error occurred when converting to float64 - %e", err),
				http.StatusInternalServerError,
			)
			return
		}
	}

	if strings.ToUpper(updateMetric.MType) == "COUNTER" {
		err := c.MemStore.SetMetric(updateMetric.MType, updateMetric.ID, updateMetric.Delta)
		if err != nil {
			c.Logger.Error().Err(err).Msg("error of SetMetric")
			http.Error(
				res,
				fmt.Sprintf("Error occurred when converting to int64 - %e", err),
				http.StatusInternalServerError,
			)
			return
		}
		res.Header().Add("content-type", "application/json")
		res.Header().Add("Date", time.Now().Format(http.TimeFormat))
		res.WriteHeader(200)
	}
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

func (c *serverController) getAllStats(res http.ResponseWriter, _ *http.Request) {
	allStats := c.MemStore.GetAllData()
	res.Header().Add("Content-Type", "application/json")
	if _, err := res.Write([]byte(allStats)); err != nil {
		c.Error().Err(err)
	}
}

func (c *serverController) Route() *chi.Mux {
	router := chi.NewRouter()
	router.Use(c.reqResLogging)
	router.Route("/", func(router chi.Router) {
		router.Get("/", c.getAllStats)
		router.Get("/value", c.returnMetric)
		router.Post("/update", c.getMetrics)
	})
	return router
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
