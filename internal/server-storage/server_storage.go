package serverstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

type Gauge float64
type Counter int64

type MemStorage struct {
	Gauge     map[string]Gauge
	Counter   map[string]Counter
	PollCount Counter
	IOHandler
}

type Dump struct {
}

type CLogger interface {
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
}

type IOHandler interface {
	WriteDump(dumpPath string, log CLogger, storage *MemStorage) error
	ReadDump(dumpPath string, recover bool, log CLogger, storage *MemStorage) error
}

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

type MemStorekeeper interface {
	SetMetric(metricType string, metricName string, metricValue interface{}) error
	GetValueByName(metricType string, metricName string) (string, error)
	GetAllData() string
	IOHandler
}

type UpdateParse struct {
	MetricType string
	MetricName string
	MetricVal  string
}

func (m *MemStorage) SetMetric(metricType string, metricName string, metricValue interface{}) error {
	if strings.ToUpper(metricType) == GaugeStr {
		m.PollCount += 1
		if fmt.Sprintf("%T", metricValue) == "string" {
			flt, err := strconv.ParseFloat(metricValue.(string), 64)
			if err != nil {
				return fmt.Errorf("can't convert to float64 %e", err)
			}
			m.Gauge[metricName] = Gauge(flt)
			return nil
		}
		m.Gauge[metricName] = Gauge(reflect.ValueOf(metricValue).Elem().Float())
		return nil

	}
	if strings.ToUpper(metricType) == CounterStr {
		m.PollCount += 1
		if _, ok := m.Counter[metricName]; !ok {
			if fmt.Sprintf("%T", metricValue) == "string" {
				integ, err := strconv.ParseInt(metricValue.(string), 10, 64)
				if err != nil {
					return fmt.Errorf("can't convert to int64 %e", err)
				}
				m.Counter[metricName] = Counter(integ)
				return nil
			}
			m.Counter[metricName] = Counter(reflect.ValueOf(metricValue).Elem().Int())
			return nil
		} else {
			if fmt.Sprintf("%T", metricValue) == "string" {
				integ, err := strconv.ParseInt(metricValue.(string), 10, 64)
				if err != nil {
					return fmt.Errorf("can't convert to int64 %e", err)
				}
				m.Counter[metricName] += Counter(integ)
				return nil
			}
			m.Counter[metricName] += Counter(reflect.ValueOf(metricValue).Elem().Int())
			return nil
		}
	}
	return fmt.Errorf("metric with type %s not found", metricType)
}

func (m *MemStorage) GetValueByName(metricType string, metricName string) (string, error) {
	if strings.ToUpper(metricType) == GaugeStr {
		if val, ok := m.Gauge[metricName]; !ok {
			return "", fmt.Errorf("there's no gauge metric called %s", metricName)
		} else {
			return strconv.FormatFloat(float64(val), 'f', -1, 64), nil
		}
	} else if strings.ToUpper(metricType) == CounterStr {
		if val, ok := m.Counter[metricName]; !ok {
			return "", fmt.Errorf("there's no counter metric called %s", metricName)
		} else {
			return strconv.FormatInt(int64(val), 10), nil
		}
	}
	return "", fmt.Errorf("metric with type %s not found", metricType)
}

func (m *MemStorage) GetAllData() string {
	allStats := []string{}
	for key, val := range m.Gauge {
		allStats = append(allStats, fmt.Sprintf("%s: %s", key, fmt.Sprintf("%f", val)))
	}
	for key, val := range m.Counter {
		allStats = append(allStats, fmt.Sprintf("%s: %s", key, fmt.Sprintf("%d", val)))
	}
	return strings.Join(allStats, "\n")
}

func NewMemStorage(dump *Dump) *MemStorage {
	return &MemStorage{
		Gauge:     map[string]Gauge{},
		Counter:   map[string]Counter{},
		PollCount: 0,
		IOHandler: dump,
	}
}

func (d Dump) ReadDump(dumpPath string, recover bool, log CLogger, storage *MemStorage) error {
	_, err := os.Stat(dumpPath)
	if recover && err == nil {
		mute := &sync.Mutex{}
		mute.Lock()
		file, err := os.OpenFile(dumpPath, os.O_RDONLY|os.O_CREATE, 0777)
		if err != nil {
			return fmt.Errorf("can't open dump file - %e", err)
		}
		scanner := bufio.NewScanner(file)
		strs := []string{}
		for scanner.Scan() {
			strs = append(strs, scanner.Text())
		}
		err = json.Unmarshal([]byte(strs[len(strs)-2]), storage)
		if err != nil {
			return fmt.Errorf("can't unmarshal dump file - %e", err)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("can't close dump file - %e", err)
		}
		log.Info().Msg(fmt.Sprintf("metrics recieved from file %s", dumpPath))
		mute.Unlock()
		return nil
	}
	log.Info().Msg(fmt.Sprintf("nothing to recieve from file %s", dumpPath))
	return nil
}
func (d Dump) WriteDump(dumpPath string, log CLogger, storage *MemStorage) error {
	mute := &sync.Mutex{}
	mute.Lock()
	jsonMemStore, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("can't marshal dump data - %e", err)
	}
	pathSliceToFile := strings.Split(dumpPath, "/")
	if len(pathSliceToFile) > 1 {
		pathSliceToFile = pathSliceToFile[1 : len(pathSliceToFile)-1]
		err = os.MkdirAll("/"+strings.Join(pathSliceToFile, "/"), 0777)
		if err != nil {
			return fmt.Errorf("can't make dir(s) - %e", err)
		}
	}
	file, err := os.OpenFile(dumpPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
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
	log.Info().Msg(fmt.Sprintf("metrics dumped to file %s", dumpPath))
	mute.Unlock()
	return nil
}

func NewDump() *Dump {
	return &Dump{}
}
