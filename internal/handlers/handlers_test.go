package handlers

import (
	"fmt"
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
		//res  string
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
				//res:  "",
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
				//res:  "Wrong metric type or empty value\n",
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
				//res:  "Bad url\n",
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
				//res:  "Can't parse metric value\n",
			},
		},
		{
			name: "Test null value 400",
			args: args{
				url:       "/update/gauge/Alloc/",
				reqMethod: http.MethodPost,
			},
			want: want{
				code: 400,
				//res:  "Wrong metric type or empty value\n",
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
				//res:  "Method not allowed\n",
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
			//recorder := httptest.NewRecorder()
			//res := recorder.Result()
			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(res)
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			fmt.Println(string(body))
			assert.Equal(t, res.StatusCode, tt.want.code)
			//assert.Equal(t, string(body), tt.want.res)
		},
		)
	}
}
