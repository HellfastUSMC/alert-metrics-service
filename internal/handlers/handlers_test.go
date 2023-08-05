package handlers

import (
	"fmt"
	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMetrics(t *testing.T) {
	type args struct {
		url       string
		reqMethod string
	}
	type want struct {
		code int
	}
	router := chi.NewRouter()

	router.Route("/update", func(router chi.Router) {
		router.Post("/{metricType}/{metricName}/{metricValue}", GetMetrics)
	})

	ts := httptest.NewServer(router)
	defer ts.Close()
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
			},
			want: want{
				code: 405,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				fmt.Println(err)
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
	}
	type want struct {
		code int
	}
	router := chi.NewRouter()
	router.Route("/", func(router chi.Router) {
		router.Get("/", GetAllStats)
	})
	ts := httptest.NewServer(router)
	defer ts.Close()
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
			},
			want: want{
				code: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
				fmt.Println(err)
			}
			assert.Equal(t, res.StatusCode, tt.want.code)
		})
	}
}

func TestReturnMetric(t *testing.T) {
	type args struct {
		url         string
		reqMethod   string
		metricName  string
		metricValue storage.Gauge
	}
	type want struct {
		code     int
		body     string
		wantBody bool
	}
	router := chi.NewRouter()
	router.Route("/value", func(router chi.Router) {
		router.Get("/{metricType}/{metricName}", ReturnMetric)
	})
	ts := httptest.NewServer(router)
	defer ts.Close()
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
			},
			want: want{
				code:     http.StatusOK,
				body:     "testMetric: 100.000000",
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
			},
			want: want{
				code:     http.StatusNotFound,
				wantBody: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Store.Gauge[tt.args.metricName] = tt.args.metricValue
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
				fmt.Println(err)
			}
			body, _ := io.ReadAll(res.Body)
			defer res.Body.Close()
			assert.Equal(t, res.StatusCode, tt.want.code)
			if tt.want.wantBody {
				assert.Equal(t, string(body), tt.want.body)
			}
		})
	}
}
