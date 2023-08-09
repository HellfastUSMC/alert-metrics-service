package controllers

import "github.com/HellfastUSMC/alert-metrics-service/internal/config"

type AgentController struct {
	Logger  CLogger
	Config  *config.SysConfig
	Storage AgentHandler
}
