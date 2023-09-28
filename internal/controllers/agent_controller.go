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

func NewAgentController(logger logger.CLogger, conf *config.SysConfig, agentHndl agentHandler) *agentController {
	return &agentController{
		Logger:  logger,
		Config:  conf,
		Storage: agentHndl,
	}
}

func (c *agentController) RenewMetrics() {
	c.Storage.RenewMetrics()
}

func (c *agentController) RenewMemCPUMetrics() {
	c.Storage.RenewMemCPUMetrics()
}

func (c *agentController) SendMetrics(key string, url string) error {
	err := c.Storage.SendBatchMetrics(key, url)
	if err != nil {
		return err
	}
	return nil
}
