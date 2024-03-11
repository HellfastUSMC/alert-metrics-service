package utils

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

func Worker(
	retryFunc func(logger logger.CLogger, intervals []int, errorsToRetry []any, function func() error) error,
	sendMetrics func(keyPath string, serverAddress string) error, log logger.CLogger,
	errorsList []any, intervals []int, keyPath string, serverAddress string,
) error {
	if err1 := sendMetrics(keyPath, "http://"+serverAddress); err1 != nil {
		log.Error().Err(err1).Msg("Error when sending metrics to server")
		f := func() error {
			err2 := sendMetrics(keyPath, "http://"+serverAddress)
			if err2 != nil {
				return err2
			}
			return nil
		}
		err := retryFunc(log, intervals, errorsList, f)
		log.Error().Err(err).Msg(fmt.Sprintf("Error after %d retries", len(intervals)+1))
	}
	return nil
}
