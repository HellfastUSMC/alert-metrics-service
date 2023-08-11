package controllers

import "github.com/HellfastUSMC/alert-metrics-service/internal/config"

type agentController struct {
	Logger  CLogger
	Config  *config.SysConfig
	Storage agentHandler
}

func (c *agentController) Info(i interface{}) {
	c.Logger.Info(i)
}

func (c *agentController) Warn(i interface{}) {
	c.Logger.Warn(i)
}

func (c *agentController) Warning(i interface{}) {
	c.Logger.Warning(i)
}

func (c *agentController) Error(i interface{}) {
	c.Logger.Error(i)
}

func (c *agentController) Infof(s string, args ...interface{}) {
	c.Logger.Infof(s, args)
}

func (c *agentController) Warnf(s string, args ...interface{}) {
	c.Logger.Warnf(s, args)
}

func (c *agentController) Errorf(s string, args ...interface{}) {
	c.Logger.Errorf(s, args)
}

func NewAgentController(logger CLogger, conf *config.SysConfig, agentHndl agentHandler) *agentController {
	return &agentController{
		Logger:  logger,
		Config:  conf,
		Storage: agentHndl,
	}
}
