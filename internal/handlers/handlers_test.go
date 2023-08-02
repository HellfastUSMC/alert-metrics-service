package handlers

import (
	"fmt"
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
		res  string
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
			},
			want: want{
				code: 200,
				res:  "",
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
				res:  "Wrong metric type or empty value\n",
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
				res:  "Bad url\n",
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
				res:  "Can't parse metric value\n",
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
				res:  "Wrong metric type or empty value\n",
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
				res:  "Method not allowed\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(
				tt.args.reqMethod,
				tt.args.url,
				nil,
			)
			if err != nil {
				t.Error(err)
			}
			//fmt.Println(tt.args.baseUrl + tt.args.metricType + tt.args.metricName + tt.args.metricValue)
			recorder := httptest.NewRecorder()
			GetMetrics(recorder, req)
			res := recorder.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			fmt.Println(string(body), res.Header.Get("val"), res.Header.Get("name"), res.Header.Get("type"), res.Header.Get("url"))
			assert.Equal(t, res.StatusCode, tt.want.code)
			assert.Equal(t, string(body), tt.want.res)
		},
		)
	}
}
