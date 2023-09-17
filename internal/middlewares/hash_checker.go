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
				headerHash := req.Header.Get("HashSHA256")
				body, err := io.ReadAll(req.Body)
				if err != nil {
					log.Error().Err(err).Msg("")
				}
				fmt.Println(string(body))
				hash := sha256.New()
				hash.Write(body)
				hashEncoded := make([]byte, 64)
				hex.Encode(hashEncoded, hash.Sum(nil))
				req.ContentLength = int64(len(body))
				req.Body = io.NopCloser(bytes.NewBuffer(body))
				fmt.Println(headerHash, string(hashEncoded))
				if string(hashEncoded) == headerHash {
					h.ServeHTTP(res, req)
				} else {
					log.Error().Err(err).Msg(fmt.Sprintf("Hash not equal header hash - %s, calculated hash - %s", headerHash, string(hashEncoded)))
					http.Error(res, "Hash not equal", http.StatusInternalServerError)
				}
			}
		})
	}
}
