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
	Storage *serverstorage.MemStorage
}

func (fd FileDump) ReadDump() error {
	_, err := os.Stat(fd.Path)
	if fd.Recover && err == nil {
		mute := &sync.Mutex{}
		mute.Lock()
		file, err := os.OpenFile(fd.Path, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			return fmt.Errorf("can't open dump file - %e", err)
		}
		scanner := bufio.NewScanner(file)
		strs := []string{}
		for scanner.Scan() {
			strs = append(strs, scanner.Text())
		}
		err = json.Unmarshal([]byte(strs[len(strs)-2]), fd.Storage)
		if err != nil {
			return fmt.Errorf("can't unmarshal dump file - %e", err)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("can't close dump file - %e", err)
		}
		log.Info().Msg(fmt.Sprintf("metrics recieved from file %s", fd.Path))
		mute.Unlock()
		return nil
	}
	log.Info().Msg(fmt.Sprintf("nothing to recieve from file %s", fd.Path))
	return nil
}

func (fd FileDump) WriteDump() error {
	mute := &sync.Mutex{}
	mute.Lock()
	jsonMemStore, err := json.Marshal(fd.Storage)
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

func NewFileDump(filePath string, recover bool, logger controllers.CLogger, storage *serverstorage.MemStorage) *FileDump {
	return &FileDump{
		Path:    filePath,
		Recover: recover,
		Logger:  logger,
		Storage: storage,
	}
}
