// Package config Пакет для конфигурирования агента и сервера при их запуске, содержит в себе структуру конфигурации,
// методы ее создания и работы с ней
package config

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"
	"runtime"

	"github.com/caarlos0/env/v6"
)

// SysConfig Структура конфигурации с указанием названий переменных окружения
type SysConfig struct {
	ServerAddress  string `env:"ADDRESS" json:"address"`
	DBPath         string `env:"DATABASE_DSN" json:"database_dsn"`
	DumpPath       string `env:"FILE_STORAGE_PATH" json:"store_file"`
	PollInterval   int64  `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval int64  `env:"REPORT_INTERVAL" json:"report_interval"`
	StoreInterval  int64  `env:"STORE_INTERVAL" json:"store_interval"`
	Recover        bool   `env:"RESTORE" json:"restore"`
	RateLimit      int64  `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key"`
	FileConfigPath string `env:"CONFIG"`
	KeyPath        string `env:"KEY_PATH" json:"key_path"`
}

// ParseServerFlags Функция парсинга флагов при запуске сервера
func (c *SysConfig) ParseServerFlags() error {
	serverFlags := flag.NewFlagSet("server config", flag.ExitOnError)
	serverFlags.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server string")
	serverFlags.StringVar(&c.DumpPath, "f", "/tmp/metrics-db.json", "Path to dump file string")
	serverFlags.Int64Var(&c.StoreInterval, "i", 300, "Storing interval in seconds int")
	serverFlags.BoolVar(&c.Recover, "r", true, "Recover from file sign bool")
	serverFlags.StringVar(
		&c.DBPath,
		"d",
		"",
		"DB connection string",
	)
	serverFlags.StringVar(&c.CryptoKey, "crypto-key", "", "Key string")
	serverFlags.StringVar(&c.FileConfigPath, "config", "", "Config file path")
	if err := serverFlags.Parse(os.Args[1:]); err != nil {
		runtime.Goexit()
		return err
	}
	return nil
}

// ParseAgentFlags Функция парсинга флагов при запуске агента
func (c *SysConfig) ParseAgentFlags() error {
	agentFlags := flag.NewFlagSet("agent config", flag.ExitOnError)
	agentFlags.StringVar(&c.ServerAddress, "a", "localhost:8080", "Address and port of server")
	agentFlags.Int64Var(&c.ReportInterval, "r", 2, "Report interval in seconds")
	agentFlags.Int64Var(&c.PollInterval, "p", 10, "Metric poll interval in seconds")
	agentFlags.Int64Var(&c.RateLimit, "l", 1, "Rate limit int")
	agentFlags.StringVar(&c.CryptoKey, "crypto-key", "", "Key string")
	agentFlags.StringVar(&c.FileConfigPath, "config", "", "Config file path")
	agentFlags.StringVar(&c.KeyPath, "k", "", "key file path")
	if err := agentFlags.Parse(os.Args[1:]); err != nil {
		runtime.Goexit()
		return err
	}
	return nil
}

// NewConfig Функция инициализации новой структуры конфигурации
func NewConfig() (*SysConfig, error) {
	config := SysConfig{}
	return &config, nil
}

// GetAgentConfigData Функция парсинга глобальных переменных/флагов/файла конфигурации для агента
func GetAgentConfigData() (*SysConfig, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, err
	}
	err = conf.ParseAgentFlags()
	if err != nil {
		return nil, err
	}
	if err = env.Parse(conf); err != nil {
		return conf, err
	}
	if conf.FileConfigPath != "" {
		conf, err = readConfFile(conf.FileConfigPath, conf)
	}
	return conf, nil
}

// GetServerConfigData Функция парсинга глобальных переменных/флагов/файла конфигурации для сервера
func GetServerConfigData() (*SysConfig, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, err
	}
	err = conf.ParseServerFlags()
	if err != nil {
		return nil, err
	}
	if err := env.Parse(conf); err != nil {
		return conf, err
	}
	if conf.FileConfigPath != "" {
		conf, err = readConfFile(conf.FileConfigPath, conf)
	}
	return conf, nil
}

func readConfFile(confFilePath string, config *SysConfig) (*SysConfig, error) {
	confFile, err := os.OpenFile(confFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(confFile)
	var strs []string
	for scanner.Scan() {
		strs = append(strs, scanner.Text())
	}
	err = confFile.Close()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(strs[len(strs)-2]), config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
