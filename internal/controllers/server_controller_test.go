package controllers

import (
	"bytes"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/HellfastUSMC/alert-metrics-service/internal/config"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	type args struct {
		url       string
		reqMethod string
		ctrl      *serverController
	}
	type want struct {
		code int
	}

	log := zerolog.New(os.Stdout)
	conf, _ := config.NewConfig()
	mStore := serverstorage.NewMemStorage(nil, &log)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test gauge 200 OK",
			args: args{
				url:       "/update/gauge/Alloc/777.5",
				reqMethod: http.MethodPost,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: 200,
			},
		},
		{
			name: "Test wrong metric type 400",
			args: args{
				url:       "/update/ga1uge/Alloc/777.5",
				reqMethod: http.MethodPost,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "Test wrong url 404",
			args: args{
				url:       "/update/gauge/777.5",
				reqMethod: http.MethodPost,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "Test wrong value 400",
			args: args{
				url:       "/update/gauge/Alloc/agb",
				reqMethod: http.MethodPost,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "Test null value 404",
			args: args{
				url:       "/update/gauge/Alloc/",
				reqMethod: http.MethodPost,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: 404,
			},
		},
		{
			name: "Test method 405",
			args: args{
				url:       "/update/gauge/Alloc/777.5",
				reqMethod: http.MethodGet,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: 405,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()

			router.Route("/update", func(router chi.Router) {
				router.Post("/{metricType}/{metricName}/{metricValue}", tt.args.ctrl.getMetrics)
			})

			ts := httptest.NewServer(router)
			defer ts.Close()
			client := ts.Client()
			req, err := http.NewRequest(
				tt.args.reqMethod,
				ts.URL+tt.args.url,
				nil,
			)
			if err != nil {
				t.Error(err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}
			assert.Equal(t, res.StatusCode, tt.want.code)
		},
		)
	}
}

func TestGetAllStats(t *testing.T) {
	type args struct {
		url       string
		reqMethod string
		ctrl      *serverController
	}
	type want struct {
		code int
	}

	log := zerolog.New(os.Stdout)
	conf, _ := config.NewConfig()
	mStore := serverstorage.NewMemStorage(nil, &log)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test 200",
			args: args{
				url:       "/",
				reqMethod: http.MethodGet,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Test 404",
			args: args{
				url:       "/444",
				reqMethod: http.MethodGet,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Test wrong method 400",
			args: args{
				url:       "/",
				reqMethod: http.MethodPost,
				ctrl:      NewServerController(&log, conf, mStore),
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Route("/", func(router chi.Router) {
				router.Get("/", tt.args.ctrl.getAllStats)
			})
			ts := httptest.NewServer(router)
			defer ts.Close()
			client := ts.Client()
			req, err := http.NewRequest(
				tt.args.reqMethod,
				ts.URL+tt.args.url,
				nil,
			)
			if err != nil {
				t.Error(err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}
			assert.Equal(t, res.StatusCode, tt.want.code)
		})
	}
}

func TestReturnMetric(t *testing.T) {
	type args struct {
		url         string
		reqMethod   string
		ctrl        *serverController
		metricName  string
		metricValue serverstorage.Gauge
	}
	type want struct {
		code     int
		body     string
		wantBody bool
	}

	log := zerolog.New(os.Stdout)
	conf, _ := config.NewConfig()
	mStore := serverstorage.NewMemStorage(nil, &log)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test 200",
			args: args{
				url:         "/value/gauge/testMetric",
				reqMethod:   http.MethodGet,
				metricName:  "testMetric",
				metricValue: 100,
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusOK,
				body:     "100",
				wantBody: true,
			},
		},
		{
			name: "Test wrong metric name",
			args: args{
				url:         "/value/gauge/test1Metric",
				reqMethod:   http.MethodGet,
				metricName:  "testMetric",
				metricValue: 100,
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusNotFound,
				wantBody: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Route("/value", func(router chi.Router) {
				router.Get("/{metricType}/{metricName}", tt.args.ctrl.returnMetric)
			})
			ts := httptest.NewServer(router)
			defer ts.Close()
			if err := tt.args.ctrl.MemStore.SetMetric(
				"gauge",
				tt.args.metricName,
				fmt.Sprintf("%v",
					tt.args.metricValue,
				),
			); err != nil {
				t.Error(err)
			}
			client := ts.Client()
			req, err := http.NewRequest(
				tt.args.reqMethod,
				ts.URL+tt.args.url,
				nil,
			)
			if err != nil {
				t.Error(err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}
			body, _ := io.ReadAll(res.Body)
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}
			assert.Equal(t, res.StatusCode, tt.want.code)
			if tt.want.wantBody {
				assert.Equal(t, string(body), tt.want.body)
			}
		})
	}
}

func Test_returnJSONMetric(t *testing.T) {
	type args struct {
		url         string
		reqMethod   string
		ctrl        *serverController
		metricName  string
		metricValue any
		metricType  string
	}
	type want struct {
		code     int
		body     string
		wantBody bool
	}

	log := zerolog.New(os.Stdout)
	conf, _ := config.NewConfig()
	mStore := serverstorage.NewMemStorage(nil, &log)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test 200 Gauge",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testGaugeMetric",
				metricType:  "gauge",
				metricValue: serverstorage.Gauge(100),
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusOK,
				body:     "{\"id\":\"testGaugeMetric\",\"type\":\"gauge\",\"value\":100}",
				wantBody: true,
			},
		},
		{
			name: "Test 200 Counter",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testCounterMetric",
				metricValue: serverstorage.Counter(100),
				metricType:  "counter",
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusOK,
				body:     "{\"id\":\"testCounterMetric\",\"type\":\"counter\",\"delta\":100}",
				wantBody: true,
			},
		},
		{
			name: "Test wrong metric name",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testMetric",
				metricType:  "gauge",
				metricValue: 100,
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusNotFound,
				wantBody: false,
			},
		},
		{
			name: "Test wrong metric type",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testMetric",
				metricType:  "gau1ge",
				metricValue: 100,
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusBadRequest,
				wantBody: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Route("/", func(router chi.Router) {
				router.Post(tt.args.url, tt.args.ctrl.returnJSONMetric)
			})
			ts := httptest.NewServer(router)
			defer ts.Close()
			if tt.want.code != http.StatusNotFound && tt.want.code != http.StatusBadRequest {
				if err := tt.args.ctrl.MemStore.SetMetric(
					tt.args.metricType,
					tt.args.metricName,
					fmt.Sprintf("%v",
						tt.args.metricValue,
					),
				); err != nil {
					t.Error(err)
				}
			}
			client := ts.Client()
			bodyString := fmt.Sprintf("{\"id\": \"%s\",\n  \"type\": \"%s\"}", tt.args.metricName, tt.args.metricType)
			buf := bytes.NewBuffer([]byte(bodyString))
			req, err := http.NewRequest(
				tt.args.reqMethod,
				ts.URL+tt.args.url,
				buf,
			)
			if err != nil {
				t.Error(err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}
			body, _ := io.ReadAll(res.Body)
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}
			log.Info().Msg(res.Status)
			assert.Equal(t, tt.want.code, res.StatusCode)
			if tt.want.wantBody {
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}

func Test_getJSONMetrics(t *testing.T) {
	type args struct {
		url         string
		reqMethod   string
		ctrl        *serverController
		metricName  string
		metricValue any
		metricType  string
	}
	type want struct {
		code     int
		body     string
		wantBody bool
	}

	log := zerolog.New(os.Stdout)
	conf, _ := config.NewConfig()
	mStore := serverstorage.NewMemStorage(nil, &log)

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test 200 Gauge",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testGaugeMetric",
				metricType:  "gauge",
				metricValue: serverstorage.Gauge(100),
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusOK,
				body:     "{\"id\":\"testGaugeMetric\",\"type\":\"gauge\",\"value\":100}",
				wantBody: true,
			},
		},
		{
			name: "Test 200 Counter",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testCounterMetric",
				metricValue: serverstorage.Counter(100),
				metricType:  "counter",
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusOK,
				body:     "{\"id\":\"testCounterMetric\",\"type\":\"counter\",\"delta\":100}",
				wantBody: true,
			},
		},
		{
			name: "Test wrong metric name",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testMetric",
				metricType:  "gauge",
				metricValue: 100,
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusNotFound,
				wantBody: false,
			},
		},
		{
			name: "Test wrong metric type",
			args: args{
				url:         "/value/",
				reqMethod:   http.MethodPost,
				metricName:  "testMetric",
				metricType:  "gau1ge",
				metricValue: 100,
				ctrl:        NewServerController(&log, conf, mStore),
			},
			want: want{
				code:     http.StatusBadRequest,
				wantBody: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Route("/", func(router chi.Router) {
				router.Post(tt.args.url, tt.args.ctrl.getMetrics)
			})
			ts := httptest.NewServer(router)
			defer ts.Close()
			if tt.want.code != http.StatusNotFound && tt.want.code != http.StatusBadRequest {
				if err := tt.args.ctrl.MemStore.SetMetric(
					tt.args.metricType,
					tt.args.metricName,
					fmt.Sprintf("%v",
						tt.args.metricValue,
					),
				); err != nil {
					t.Error(err)
				}
			}
			client := ts.Client()
			num := "delta"
			if tt.args.metricType == "gauge" {
				num = "value"
			}
			bodyString := fmt.Sprintf("{\"id\": \"%s\",\n  \"type\": \"%s\",\n  \"%s\": \"%v\"}", tt.args.metricName, tt.args.metricType, num, tt.args.metricValue)
			buf := bytes.NewBuffer([]byte(bodyString))
			req, err := http.NewRequest(
				tt.args.reqMethod,
				ts.URL+tt.args.url,
				buf,
			)
			if err != nil {
				t.Error(err)
			}
			res, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}
			body, _ := io.ReadAll(res.Body)
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}
			log.Info().Msg(res.Status)
			assert.Equal(t, tt.want.code, res.StatusCode)
			if tt.want.wantBody {
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}
