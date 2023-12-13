// Package config Пакет для конфигурирования агента и сервера при их запуске, содержит в себе структуру конфигурации,
// методы ее создания и работы с ней
package config

import (
	"flag"
	"os"
	"runtime"

	"github.com/caarlos0/env/v6"
)

// SysConfig Структура конфигурации с указанием названий переменных окружения
type SysConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	DBPath         string `env:"DATABASE_DSN"`
	DumpPath       string `env:"FILE_STORAGE_PATH"`
	Key            string `env:"KEY"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	StoreInterval  int64  `env:"STORE_INTERVAL"`
	Recover        bool   `env:"RESTORE"`
	RateLimit      int64  `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	CryptoCert     string `env:"CRYPTO_CERT"`
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
	serverFlags.StringVar(&c.Key, "k", "", "Hash key string")
	serverFlags.StringVar(&c.CryptoKey, "crypto-key", "", "Key string")
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
	agentFlags.StringVar(&c.Key, "k", "", "Hash key string")
	agentFlags.Int64Var(&c.RateLimit, "l", 1, "Rate limit int")
	agentFlags.StringVar(&c.CryptoKey, "crypto-key", "", "Key string")
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

// GetAgentConfigData Функция парсинга глобальных переменных и флагов для агента
func GetAgentConfigData() (*SysConfig, error) {
	conf, err := NewConfig()
	if err != nil {
		return nil, err
	}
	err = conf.ParseAgentFlags()
	if err != nil {
		return nil, err
	}
	if err := env.Parse(conf); err != nil {
		return conf, err
	}
	return conf, nil
}

// GetServerConfigData Функция парсинга глобальных переменных и флагов для сервера
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
	return conf, nil
}
