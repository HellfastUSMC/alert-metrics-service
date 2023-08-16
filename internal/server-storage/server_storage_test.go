package serverstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_SetMetric(t *testing.T) {
	type fields struct {
		Metrics   map[string]Gauge
		PollCount Counter
	}
	type args struct {
		metricName  string
		metricValue string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test normal value",
			fields: fields{
				Metrics:   map[string]Gauge{},
				PollCount: 0,
			},
			args: args{
				metricName:  "testMetric",
				metricValue: "777.5",
			},
			wantErr: false,
		},
		{
			name: "Test rune value",
			fields: fields{
				Metrics:   map[string]Gauge{},
				PollCount: 0,
			},
			args: args{
				metricName:  "testMetric",
				metricValue: "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauge:     tt.fields.Metrics,
				PollCount: tt.fields.PollCount,
			}
			if err := m.SetMetric("Gauge", tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("SetMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_GetValueByName(t *testing.T) {
	type fields struct {
		Metrics map[string]Gauge
	}
	type args struct {
		metricName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Test normal behaviour",
			fields:  fields{Metrics: map[string]Gauge{"Alloc": 797.5}},
			args:    args{metricName: "Alloc"},
			want:    "797.5",
			wantErr: false,
		},
		{
			name:    "Test error behaviour",
			fields:  fields{Metrics: map[string]Gauge{"Alloc": 797.5}},
			args:    args{metricName: "ABC"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauge: tt.fields.Metrics,
			}
			got, err := m.GetValueByName("Gauge", tt.args.metricName)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equalf(t, tt.want, got, "GetValueByName(%v)", tt.args.metricName)
		})
	}
}
