// Package middlewares Пакет мидлварей используемых в работе серверной части приложения - проверка хэша,
// логгирование запросов, обработка GZip
package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

// CheckCert Мидлварь для проверки шифрования запроса
func CheckCert(log logger.CLogger, privateKeyPath string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			privateKey, err := os.ReadFile(privateKeyPath)
			if err != nil {
				http.Error(res, "Cannot get private key", http.StatusInternalServerError)
				return
			}
			privateKeyBlock, _ := pem.Decode(privateKey)
			privateKeyBytes, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
			if err != nil {
				http.Error(res, "Cannot parse private key", http.StatusInternalServerError)
				return
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				log.Error().Err(err).Msg("")
			}
			decodedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKeyBytes, body)
			if err != nil {
				http.Error(res, "Cannot decode data with private key", http.StatusInternalServerError)
				return
			}
			req.ContentLength = int64(len(decodedData))
			req.Body = io.NopCloser(bytes.NewBuffer(decodedData))
			h.ServeHTTP(res, req)
		})
	}
}
