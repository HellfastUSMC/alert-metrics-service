package controllers

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
)

type agentController struct {
	Logger  CLogger
	Config  *config.SysConfig
	Storage agentHandler
}

func NewAgentController(logger CLogger, conf *config.SysConfig, agentHndl agentHandler) *agentController {
	return &agentController{
		Logger:  logger,
		Config:  conf,
		Storage: agentHndl,
	}
}

func (c *agentController) RenewMetrics() {
	c.Storage.RenewMetrics()
}

func (c *agentController) SendMetrics(url string) error {
	err := c.Storage.SendMetrics(url)
	if err != nil {
		return err
	}
	return nil
}
