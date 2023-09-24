package middlewares

import (
	"bytes"
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/utils"
	"io"
	"net/http"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

type CHRespWriter struct {
	http.ResponseWriter
}

func CheckHash(log logger.CLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if req.Header.Get("HashSHA256") != "" {
				headerHash := req.Header.Get("HashSHA256")
				body, err := io.ReadAll(req.Body)
				if err != nil {
					log.Error().Err(err).Msg("")
				}
				//hash := sha256.New()
				//hash.Write(body)
				//hashEncoded := make([]byte, 64)
				//hex.Encode(hashEncoded, hash.Sum(nil))
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
