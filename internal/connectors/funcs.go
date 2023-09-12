package connectors

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
)

func GetDumper(log logger.CLogger, conf *config.SysConfig) (serverstorage.Dumper, error) {
	if conf.DBPath != "" {
		dumper, err := NewConnectionPGSQL(conf.DBPath, log)
		if err != nil || dumper == nil {
			log.Error().Err(err).Msg("error in creating new SQL connection")
			return nil, err
		}
		log.Info().Msg("Using DB dumper")
		return dumper, nil
	} else if conf.DumpPath != "" {
		dumper := NewFileDump(conf.DumpPath, conf.Recover, log)
		log.Info().Msg("Using file dumper")
		return dumper, nil
	} else {
		log.Info().Msg("Using memory dumper")
		return nil, nil
	}
}
