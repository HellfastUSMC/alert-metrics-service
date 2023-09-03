package connectors

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"github.com/rs/zerolog/log"
)

type FileDump struct {
	path    string
	recover bool
	logger  logger.CLogger
}

func (fd FileDump) ReadDump() ([]string, error) {
	_, err := os.Stat(fd.path)
	if err == nil {
		file, err := os.OpenFile(fd.path, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			return nil, fmt.Errorf("can't open dump file - %e", err)
		}
		scanner := bufio.NewScanner(file)
		strs := []string{}
		for scanner.Scan() {
			strs = append(strs, scanner.Text())
		}
		err = file.Close()
		if err != nil {
			return nil, err
		}
		return strs, nil
	}
	log.Info().Msg(fmt.Sprintf("nothing to recieve from file %s", fd.path))
	return nil, err
}

func (fd FileDump) WriteDump(jsonMemStore []byte) error {
	pathSliceToFile := strings.Split(fd.path, "/")
	if len(pathSliceToFile) > 1 {
		pathSliceToFile = pathSliceToFile[1 : len(pathSliceToFile)-1]
		err := os.MkdirAll("/"+strings.Join(pathSliceToFile, "/"), 0777)
		if err != nil {
			return fmt.Errorf("can't make dir(s) - %e", err)
		}
	}
	file, err := os.OpenFile(fd.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return fmt.Errorf("can't open a file - %e", err)
	}
	jsonMemStore = append(jsonMemStore, []byte("\n")...)
	_, err = file.Write(jsonMemStore)
	if err != nil {
		return fmt.Errorf("can't write json to a file - %e", err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("can't close a file - %e", err)
	}
	log.Info().Msg(fmt.Sprintf("metrics dumped to file %s", fd.path))
	return nil
}

func (fd FileDump) GetPath() string {
	return fd.path
}

func (fd FileDump) Ping() error {
	return nil
}

func NewFileDump(filePath string, recover bool, logger logger.CLogger) *FileDump {
	return &FileDump{
		path:    filePath,
		recover: recover,
		logger:  logger,
	}
}
