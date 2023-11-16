package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"github.com/HellfastUSMC/alert-metrics-service/internal/utils"
)

//type CHRespWriter struct {
//	http.ResponseWriter
//}

// CheckHash Мидлварь для проверки хэша запроса
func CheckHash(log logger.CLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if req.Header.Get("HashSHA256") != "" {
				headerHash := req.Header.Get("HashSHA256")
				body, err := io.ReadAll(req.Body)
				if err != nil {
					log.Error().Err(err).Msg("")
				}
				req.ContentLength = int64(len(body))
				req.Body = io.NopCloser(bytes.NewBuffer(body))

				hasher := utils.NewHasher()
				hasher.CalcHexHash(body)

				if hasher.String() == headerHash {
					h.ServeHTTP(res, req)
				} else {
					log.Error().Err(err).Msg(fmt.Sprintf("Hash not equal header hash - %s, calculated hash - %s", headerHash, hasher.String()))
					http.Error(res, "Hash not equal", http.StatusInternalServerError)
					return
				}
			}
			h.ServeHTTP(res, req)
		})
	}
}
