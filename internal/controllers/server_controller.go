package controllers

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	data *respData
	http.ResponseWriter
}

type respData struct {
	code int
	size int
}

type gzipRespWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (c *serverController) returnJSONMetric(res http.ResponseWriter, req *http.Request) {
	updateMetric := Metrics{}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Can't read request body", http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(body, &updateMetric); err != nil {
		http.Error(res, "Can't parse JSON", http.StatusInternalServerError)
		return
	}

	if strings.ToUpper(updateMetric.MType) != "GAUGE" && strings.ToUpper(updateMetric.MType) != "COUNTER" {
		http.Error(res, "Wrong metric type", http.StatusBadRequest)
		return
	}
	val, err := c.MemStore.GetValueByName(updateMetric.MType, updateMetric.ID)
	if err != nil {
		c.Error().Err(err).Msg("error of GetValueByName ")
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusNotFound)
		return
	}
	if strings.ToUpper(updateMetric.MType) == "GAUGE" {
		flVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			c.Error().Err(err)
			http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusInternalServerError)
			return
		}
		updateMetric.Value = &flVal
	} else if strings.ToUpper(updateMetric.MType) == "COUNTER" {
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			c.Error().Err(err)
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
		c.Error().Err(err)
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
	if err := json.Unmarshal(body, &updateMetric); err != nil {
		http.Error(res, "can't unmarshal JSON", http.StatusInternalServerError)
		return
	}
	if (strings.ToUpper(updateMetric.MType) != "GAUGE" &&
		strings.ToUpper(updateMetric.MType) != "COUNTER") ||
		(updateMetric.Value == nil &&
			updateMetric.Delta == nil) {
		http.Error(res, "Wrong metric type or empty value", http.StatusBadRequest)
		return
	}

	if strings.ToUpper(updateMetric.MType) == "GAUGE" {
		err := c.MemStore.SetMetric(updateMetric.MType, updateMetric.ID, updateMetric.Value)
		if err != nil {
			c.Logger.Error().Err(err).Msg("error of SetMetric")
			http.Error(
				res,
				fmt.Sprintf("Error occurred when setting metric - %e", err),
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
				fmt.Sprintf("Error occurred when setting metric - %e", err),
				http.StatusInternalServerError,
			)
			return
		}
		newMetricVal, err := c.MemStore.GetValueByName(updateMetric.MType, updateMetric.ID)
		if err != nil {
			c.Logger.Error().Err(err).Msg("can't get new metric value")
		}
		intVal, err := strconv.ParseInt(newMetricVal, 10, 64)
		if err != nil {
			c.Logger.Error().Err(err).Msg("can't parse new int64 metric value")
		}
		updateMetric.Delta = &intVal
	}
	jsonData, err := json.Marshal(updateMetric)
	if err != nil {
		http.Error(res, "can't marshal JSON", http.StatusInternalServerError)
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
	if _, err = res.Write(jsonData); err != nil {
		c.Error().Err(err)
		http.Error(res, fmt.Sprintf("there's an error %e", err), http.StatusInternalServerError)
		return
	}

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
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
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
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.Header().Add("Date", time.Now().Format(http.TimeFormat))
	res.WriteHeader(http.StatusOK)
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
	res.Header().Add("Content-Type", "text/html")
	if _, err := res.Write([]byte(allStats)); err != nil {
		c.Error().Err(err)
	}
}

func (c *serverController) Route() *chi.Mux {
	router := chi.NewRouter()
	router.Use(c.reqResLogging)
	router.Use(c.gzip)
	router.Route("/", func(router chi.Router) {
		router.Get("/", c.getAllStats)
		router.Post("/value/", c.returnJSONMetric)
		router.Post("/update/", c.getJSONMetrics)
		router.Get("/value/{metricType}/{metricName}", c.returnMetric)
		router.Post("/update/{metricType}/{metricName}/{metricValue}", c.getMetrics)
	})
	//chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
	//	fmt.Printf("[%s]: '%s' has %d middlewares\n", method, route, len(middlewares))
	//	return nil
	//})
	return router
}

func (c *serverController) reqResLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		rw := logRespWriter{
			data: &respData{
				code: 0,
				size: 0,
			},
			ResponseWriter: res,
		}

		h.ServeHTTP(&rw, r)

		duration := time.Since(start).String()

		c.Info().Str("URI", uri).Str("method", method).Str("duration", duration).Int("code", rw.data.code).Int("size", rw.data.size)
	})
}

func (r *logRespWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.data.size += size
	return size, err
}

func (r *logRespWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.data.code = statusCode
}

func (c *serverController) gzip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		//fmt.Println(c.MemStore)
		fmt.Println(req.URL.String())
		if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {

			body, err := io.ReadAll(req.Body)
			if err != nil {
				c.Error().Err(err)
			}

			var buff bytes.Buffer

			reader := flate.NewReader(bytes.NewReader(body))

			_, err = buff.ReadFrom(reader)
			if err != nil {
				c.Error().Err(err)
			}

			err = reader.Close()
			if err != nil {
				c.Error().Err(err)
			}

			req.ContentLength = int64(len(buff.Bytes()))
			req.Body = io.NopCloser(&buff)
		}

		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(res, req)
		} else if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {

			gz, err := gzip.NewWriterLevel(res, gzip.BestSpeed)

			if err != nil {
				c.Error().Err(err)
				http.Error(res, "can't compress to gzip", http.StatusInternalServerError)
				return
			}

			defer gz.Close()

			res.Header().Set("Content-Encoding", "gzip")

			h.ServeHTTP(gzipRespWriter{
				ResponseWriter: res,
				Writer:         gz,
			}, req)
		}
	})
}

func (w gzipRespWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (c *serverController) ReadDump() error {
	//fmt.Println("READ")
	_, err := os.Stat(c.Config.DumpPath)
	if c.Config.Recover && err == nil {
		file, err := os.OpenFile(c.Config.DumpPath, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			return fmt.Errorf("can't open dump file - %e", err)
		}
		offset, err := file.Seek(-701, 2)
		if err != nil {
			return fmt.Errorf("can't seek dump file - %e", err)
		}
		fileEnd := make([]byte, 700)
		_, _ = file.ReadAt(fileEnd, offset)
		lastString := []byte(strings.Split(string(fileEnd), "\n")[1])
		//fmt.Println("read...", string(lastString))
		err = json.Unmarshal(lastString, c.MemStore)
		if err != nil {
			return fmt.Errorf("can't unmarshal dump file - %e", err)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("can't close dump file - %e", err)
		}
		c.Info().Msg(fmt.Sprintf("metrics recieved from file %s", c.Config.DumpPath))
		return nil
	}
	return nil
}
func (c *serverController) WriteDump() error {
	//fmt.Println("WRITE")
	jsonMemStore, err := json.Marshal(c.MemStore)
	//fmt.Println("write...", string(jsonMemStore))
	if err != nil {
		return fmt.Errorf("can't marshal dump data - %e", err)
	}
	pathSliceToFile := strings.Split(c.Config.DumpPath, "/")
	if len(pathSliceToFile) > 1 {
		pathSliceToFile = pathSliceToFile[1 : len(pathSliceToFile)-1]
		err = os.MkdirAll("/"+strings.Join(pathSliceToFile, "/"), 0777)
		if err != nil {
			//fmt.Println(err)
			return fmt.Errorf("can't make dir(s) - %e", err)
		}
	}
	file, err := os.OpenFile(c.Config.DumpPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return fmt.Errorf("can't open a file - %e", err)
	}
	jsonMemStore = append(jsonMemStore, []byte("\n")...)
	_, err = file.Write(jsonMemStore)
	if err != nil {
		return fmt.Errorf("can't write json to a file - %e", err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("can't close a file - %e", err)
	}
	c.Info().Msg(fmt.Sprintf("metrics dumped to file %s", c.Config.DumpPath))
	return nil
}

func (c *serverController) StartServer() {
	router := chi.NewRouter()
	router.Mount("/", c.Route())
	c.Info().Msg(fmt.Sprintf(
		"Starting server at %s with store interval %ds, dump path %s and recover state is %v",
		c.Config.ServerAddress,
		c.Config.StoreInterval,
		c.Config.DumpPath,
		c.Config.Recover,
	))
	err := http.ListenAndServe(c.Config.ServerAddress, c.Route())
	if err != nil {
		c.Error().Err(err)
	}
}

func (c *serverController) StartDumping() {
	tickDump := time.NewTicker(time.Duration(c.Config.StoreInterval) * time.Second)
	go func() {
		for {
			<-tickDump.C
			if err := c.WriteDump(); err != nil {
				c.Error().Err(err)
			}
		}
	}()
}
