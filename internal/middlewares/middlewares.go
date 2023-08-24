package middlewares

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type respData struct {
	code int
	size int
}

type logRespWriter struct {
	data *respData
	http.ResponseWriter
}

type gzipRespWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

type CLogger interface {
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
}

func ReqResLogging(log CLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
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

			log.Info().Str("URI", uri).Str("method", method).Str("duration", duration).Int("code", rw.data.code).Int("size", rw.data.size)
		})
	}
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

func Gzip(log CLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {

				body, err := io.ReadAll(req.Body)
				if err != nil {
					log.Error().Err(err)
				}

				var buff bytes.Buffer

				reader := flate.NewReader(bytes.NewReader(body))

				_, err = buff.ReadFrom(reader)
				if err != nil {
					log.Error().Err(err)
				}

				err = reader.Close()
				if err != nil {
					log.Error().Err(err)
				}

				req.ContentLength = int64(len(buff.Bytes()))
				req.Body = io.NopCloser(&buff)
			}

			if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
				h.ServeHTTP(res, req)
			} else if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {

				gz, err := gzip.NewWriterLevel(res, gzip.BestSpeed)

				if err != nil {
					log.Error().Err(err)
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
}

func (w gzipRespWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
