package controllers

import (
	"net/http"
)

type Logger interface {
	Info()
	Warn()
	Debug()
}

type ServerHandler interface {
	GetMetrics(res http.ResponseWriter, req *http.Request)
	GetAllStats(res http.ResponseWriter, req *http.Request)
	ReturnMetric(res http.ResponseWriter, req *http.Request)
}

//type Router interface {
//	http.Handler
//	chi.Routes
//	Use(middlewares ...func(http.Handler) http.Handler)
//	With(middlewares ...func(http.Handler) http.Handler) Router
//	Group(fn func(r Router)) Router
//	Route(pattern string, fn func(r Router)) Router
//	Mount(pattern string, h http.Handler)
//	Handle(pattern string, h http.Handler)
//	HandleFunc(pattern string, h http.HandlerFunc)
//	Method(method string, pattern string, h http.Handler)
//	MethodFunc(method string, pattern string, h http.HandlerFunc)
//	Connect(pattern string, h http.HandlerFunc)
//	Delete(pattern string, h http.HandlerFunc)
//	Get(pattern string, h http.HandlerFunc)
//	Head(pattern string, h http.HandlerFunc)
//	Options(pattern string, h http.HandlerFunc)
//	Patch(pattern string, h http.HandlerFunc)
//	Post(pattern string, h http.HandlerFunc)
//	Put(pattern string, h http.HandlerFunc)
//	Trace(pattern string, h http.HandlerFunc)
//	NotFound(h http.HandlerFunc)
//	MethodNotAllowed(h http.HandlerFunc)
//}
//
//type Config interface {
//	PollInterval
//	ReportInterval
//	ServerAddress
//}
