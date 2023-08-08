package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/HellfastUSMC/alert-metrics-service/internal/storage"
)

func TestGetMetrics(t *testing.T) {
	type args struct {
		url         string
		reqMethod   string
		storageImpl storage.MemStorage
	}
	type want struct {
		code int
	}
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				router.Post("/{metricType}/{metricName}/{metricValue}", GetMetrics(&tt.args.storageImpl))
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
		url         string
		reqMethod   string
		storageImpl storage.MemStorage
	}
	type want struct {
		code int
	}
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				router.Get("/", GetAllStats(&tt.args.storageImpl))
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
		storageImpl storage.MemStorage
		metricName  string
		metricValue storage.Gauge
	}
	type want struct {
		code     int
		body     string
		wantBody bool
	}
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				storageImpl: storage.MemStorage{
					Gauge:     map[string]storage.Gauge{},
					Counter:   map[string]storage.Counter{},
					PollCount: 0,
				},
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
				router.Get("/{metricType}/{metricName}", ReturnMetric(&tt.args.storageImpl))
			})
			ts := httptest.NewServer(router)
			defer ts.Close()
			tt.args.storageImpl.Gauge[tt.args.metricName] = tt.args.metricValue
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
