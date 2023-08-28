package connectors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/HellfastUSMC/alert-metrics-service/internal/controllers"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/rs/zerolog/log"
)

type FileDump struct {
	Path    string
	Recover bool
	Logger  controllers.CLogger
	//Storage *serverstorage.MemStorage
}

func (fd FileDump) ReadDump() ([]string, error) {
	_, err := os.Stat(fd.Path)
	if fd.Recover && err == nil {
		file, err := os.OpenFile(fd.Path, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			return nil, fmt.Errorf("can't open dump file - %e", err)
		}
		scanner := bufio.NewScanner(file)
		strs := []string{}
		for scanner.Scan() {
			strs = append(strs, scanner.Text())
		}
		err = file.Close()
		return strs, nil
	}
	log.Info().Msg(fmt.Sprintf("nothing to recieve from file %s", fd.Path))
	return nil, nil
}

func (fd FileDump) WriteDump(memStore *serverstorage.MemStorage) error {
	mute := &sync.Mutex{}
	mute.Lock()
	jsonMemStore, err := json.Marshal(memStore)
	if err != nil {
		return fmt.Errorf("can't marshal dump data - %e", err)
	}
	pathSliceToFile := strings.Split(fd.Path, "/")
	if len(pathSliceToFile) > 1 {
		pathSliceToFile = pathSliceToFile[1 : len(pathSliceToFile)-1]
		err = os.MkdirAll("/"+strings.Join(pathSliceToFile, "/"), 0777)
		if err != nil {
			return fmt.Errorf("can't make dir(s) - %e", err)
		}
	}
	file, err := os.OpenFile(fd.Path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
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
	log.Info().Msg(fmt.Sprintf("metrics dumped to file %s", fd.Path))
	mute.Unlock()
	return nil
}

func (fd *FileDump) GetPath() string {
	return fd.Path
}

func NewFileDump(filePath string, recover bool, logger controllers.CLogger) *FileDump {
	return &FileDump{
		Path:    filePath,
		Recover: recover,
		Logger:  logger,
	}
}
