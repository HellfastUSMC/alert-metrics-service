package handlers

import (
	"net/http"
	"testing"
)

func TestGetMetrics(t *testing.T) {
	type args struct {
		baseUrl     string
		metricType  string
		metricName  string
		metricValue string
		reqMethod   string
	}
	tests := []struct {
		name string
		args args
		want struct {
			code int64
			res  string
		}
	}{
		{
			name: "Test 200 OK",
			args: args{
				baseUrl:     "http://localhost:8080/update/",
				metricType:  "gauge/",
				metricName:  "Alloc/",
				metricValue: "789",
				reqMethod:   http.MethodPost,
			},
			want: struct {
				code int64
				res  string
			}{code: 200, res: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(
				tt.args.reqMethod,
				tt.args.baseUrl+tt.args.metricType+tt.args.metricName+tt.args.metricValue,
				nil,
			)

		})
	}
}
