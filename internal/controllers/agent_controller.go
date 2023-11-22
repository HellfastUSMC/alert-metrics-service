// Package controllers Пакет контроллеров сервера и агента, содержит в себе структуры, методы для создания и
// использования контроллеров, также интерфейсы и вспомогательные структуры
package controllers

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
)

type agentController struct {
	Logger  logger.CLogger
	Config  *config.SysConfig
	Storage agentHandler
}

// NewAgentController Функция инициализации нового контроллера агента
func NewAgentController(logger logger.CLogger, conf *config.SysConfig, agentHndl agentHandler) *agentController {
	return &agentController{
		Logger:  logger,
		Config:  conf,
		Storage: agentHndl,
	}
}

// RenewMetrics Функция представитель для обновления метрик
func (c *agentController) RenewMetrics() {
	c.Storage.RenewMetrics()
}

// RenewMemCPUMetrics Функция представитель для обновления дополнительных метрик
func (c *agentController) RenewMemCPUMetrics() {
	c.Storage.RenewMemCPUMetrics()
}

// SendMetrics Функция представитель для отправки метрик
func (c *agentController) SendMetrics(key string, url string) error {
	err := c.Storage.SendBatchMetrics(key, url)
	if err != nil {
		return err
	}
	return nil
}
