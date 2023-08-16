package controllers

type CLogger interface {
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Warning(args ...interface{})
}

type agentHandler interface {
	RenewMetrics()
	SendMetrics(string) error
}
