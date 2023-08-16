package controllers

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/rs/zerolog"
)

type agentController struct {
	Logger  CLogger
	Config  *config.SysConfig
	Storage agentHandler
}

func (c *agentController) Info() *zerolog.Event {
	return c.Logger.Info()
}

func (c *agentController) Warn() *zerolog.Event {
	return c.Logger.Warn()
}

func (c *agentController) Error() *zerolog.Event {
	return c.Logger.Error()
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
