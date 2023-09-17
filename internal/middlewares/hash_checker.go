package middlewares

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

func CheckHash(log logger.CLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if req.Header.Get("HashSHA256") != "" {
				body, err := io.ReadAll(req.Body)
				fmt.Println("body", string(body))
				if err != nil {
					log.Error().Err(err)
				}
				hash := sha256.New()
				hash.Write(body)
				hashDecoded := make([]byte, 64)
				hex.Encode(hashDecoded, hash.Sum(nil))
				req.ContentLength = int64(len(body))
				req.Body = io.NopCloser(bytes.NewBuffer(body))
				if string(hashDecoded) == req.Header.Get("HashSHA256") {
					h.ServeHTTP(res, req)
				} else {
					log.Error().Err(err).Msg("")
					http.Error(res, "Hash not equal", http.StatusInternalServerError)
				}
			}
		})
	}
}
