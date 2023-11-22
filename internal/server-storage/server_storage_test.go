package serverstorage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkMemStorage_SetMetric(b *testing.B) {
	m := &MemStorage{
		Gauge:     map[string]Gauge{},
		PollCount: 0,
	}
	b.ResetTimer()
	if err := m.SetMetric("Gauge", "Name", "100"); err != nil {
		b.Error(err)
	}
}

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

func BenchmarkMemStorage_GetValueByName(b *testing.B) {
	m := &MemStorage{
		Gauge:     map[string]Gauge{},
		PollCount: 0,
	}
	if err := m.SetMetric("Gauge", "Name", "100"); err != nil {
		b.Error(err)
	}
	b.ResetTimer()
	if _, err := m.GetValueByName("Gauge", "Name"); err != nil {
		b.Error(err)
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

func BenchmarkMemStorage_GetAllData(b *testing.B) {
	m := &MemStorage{
		Gauge: map[string]Gauge{
			"1": 1,
			"2": 2,
			"3": 3,
		},
		Counter: map[string]Counter{
			"1": 1,
			"2": 2,
			"3": 3,
		},
		PollCount: 100,
	}
	b.ResetTimer()
	m.GetAllData()
}

func TestMemStorage_GetAllData(t *testing.T) {
	type fields struct {
		Gauge     map[string]Gauge
		Counter   map[string]Counter
		PollCount Counter
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "normal behaviour",
			fields: fields{
				Gauge:     map[string]Gauge{"Alloc": 10.5},
				Counter:   map[string]Counter{"MAlloc": 10},
				PollCount: 0,
			},
			want: "Alloc: 10.500000\nMAlloc: 10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauge:     tt.fields.Gauge,
				Counter:   tt.fields.Counter,
				PollCount: tt.fields.PollCount,
			}
			assert.Equalf(t, tt.want, m.GetAllData(), "GetAllData()")
		})
	}
}

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		{
			name: "normal behaviour",
			want: &MemStorage{
				Gauge:   map[string]Gauge{},
				Counter: map[string]Counter{},
				Logger:  nil,
				Dumper:  nil,
				Mutex:   &sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewMemStorage(nil, nil), "NewMemStorage()")
		})
	}
}
