package middlewares

import (
	"net/http"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

type respData struct {
	code int
	size int
}

type logRespWriter struct {
	data *respData
	http.ResponseWriter
}

// ReqResLogging Мидлварь для логирования запросов
func ReqResLogging(log logger.CLogger) func(h http.Handler) http.Handler {
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
			duration := time.Since(start).String()
			log.Info().Str("URI", uri).Str("method", method).Str("duration", duration).Int("code", rw.data.code).Int("size", rw.data.size).Msg("")
			h.ServeHTTP(&rw, r)

		})
	}
}

//func (r *logRespWriter) Write(b []byte) (int, error) {
//	size, err := r.ResponseWriter.Write(b)
//	r.data.size += size
//	return size, err
//}
//
//func (r *logRespWriter) WriteHeader(statusCode int) {
//	r.ResponseWriter.WriteHeader(statusCode)
//	r.data.code = statusCode
//}
