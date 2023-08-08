package controllers

import (
	"github.com/HellfastUSMC/alert-metrics-service/internal/flags"
	"github.com/go-chi/chi/v5"
)

type ServerController struct {
	Logger  *Logger
	Config  *flags.SysConfig
	Handler *ServerHandler
	Router  chi.Router
}
