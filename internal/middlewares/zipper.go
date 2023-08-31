package middlewares

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

type gzipRespWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func Gzip(log logger.CLogger) func(h http.Handler) http.Handler {
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
