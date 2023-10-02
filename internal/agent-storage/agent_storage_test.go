package agentstorage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_RenewMetrics(t *testing.T) {
	type fields struct {
		Alloc         Gauge
		BuckHashSys   Gauge
		Frees         Gauge
		GCCPUFraction Gauge
		GCSys         Gauge
		HeapAlloc     Gauge
		HeapIdle      Gauge
		HeapInuse     Gauge
		HeapObjects   Gauge
		HeapReleased  Gauge
		HeapSys       Gauge
		LastGC        Gauge
		Lookups       Gauge
		MCacheInuse   Gauge
		MCacheSys     Gauge
		MSpanInuse    Gauge
		MSpanSys      Gauge
		Mallocs       Gauge
		NextGC        Gauge
		NumForcedGC   Gauge
		NumGC         Gauge
		OtherSys      Gauge
		PauseTotalNs  Gauge
		StackInuse    Gauge
		StackSys      Gauge
		Sys           Gauge
		TotalAlloc    Gauge
		PollCount     Counter
		RandomValue   Gauge
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Test metrics reading",
			fields: fields{
				Alloc:         0,
				BuckHashSys:   0,
				Frees:         0,
				GCCPUFraction: 0,
				GCSys:         0,
				HeapAlloc:     0,
				HeapIdle:      0,
				HeapInuse:     0,
				HeapObjects:   0,
				HeapReleased:  0,
				HeapSys:       0,
				LastGC:        0,
				Lookups:       0,
				MCacheInuse:   0,
				MCacheSys:     0,
				MSpanInuse:    0,
				MSpanSys:      0,
				Mallocs:       0,
				NextGC:        0,
				NumForcedGC:   0,
				NumGC:         0,
				OtherSys:      0,
				PauseTotalNs:  0,
				StackInuse:    0,
				StackSys:      0,
				Sys:           0,
				TotalAlloc:    0,
				PollCount:     0,
				RandomValue:   0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Alloc:         tt.fields.Alloc,
				BuckHashSys:   tt.fields.BuckHashSys,
				Frees:         tt.fields.Frees,
				GCCPUFraction: tt.fields.GCCPUFraction,
				GCSys:         tt.fields.GCSys,
				HeapAlloc:     tt.fields.HeapAlloc,
				HeapIdle:      tt.fields.HeapIdle,
				HeapInuse:     tt.fields.HeapInuse,
				HeapObjects:   tt.fields.HeapObjects,
				HeapReleased:  tt.fields.HeapReleased,
				HeapSys:       tt.fields.HeapSys,
				LastGC:        tt.fields.LastGC,
				Lookups:       tt.fields.Lookups,
				MCacheInuse:   tt.fields.MCacheInuse,
				MCacheSys:     tt.fields.MCacheSys,
				MSpanInuse:    tt.fields.MSpanInuse,
				MSpanSys:      tt.fields.MSpanSys,
				Mallocs:       tt.fields.Mallocs,
				NextGC:        tt.fields.NextGC,
				NumForcedGC:   tt.fields.NumForcedGC,
				NumGC:         tt.fields.NumGC,
				OtherSys:      tt.fields.OtherSys,
				PauseTotalNs:  tt.fields.PauseTotalNs,
				StackInuse:    tt.fields.StackInuse,
				StackSys:      tt.fields.StackSys,
				Sys:           tt.fields.Sys,
				TotalAlloc:    tt.fields.TotalAlloc,
				PollCount:     tt.fields.PollCount,
				RandomValue:   tt.fields.RandomValue,
			}
			m.RenewMetrics()
			assert.NotEqual(t, m.Alloc, tt.fields.Alloc)
			assert.NotEqual(t, m.HeapSys, tt.fields.HeapSys)
		})
	}
}

func TestMetrics_SendMetrics(t *testing.T) {
	type fields struct {
		Alloc Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Test normal condition",
			fields:  fields{Alloc: 777.5},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metric{
				Alloc: tt.fields.Alloc,
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			if err := m.SendBatchMetrics("", server.URL); (err != nil) != tt.wantErr {
				t.Errorf("SendMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMetricsStorage(t *testing.T) {
	tests := []struct {
		name string
		want *Metric
	}{
		{
			name: "normal behaviour",
			want: &Metric{
				Alloc:         0,
				BuckHashSys:   0,
				Frees:         0,
				GCCPUFraction: 0,
				GCSys:         0,
				HeapAlloc:     0,
				HeapIdle:      0,
				HeapInuse:     0,
				HeapObjects:   0,
				HeapReleased:  0,
				HeapSys:       0,
				LastGC:        0,
				Lookups:       0,
				MCacheInuse:   0,
				MCacheSys:     0,
				MSpanInuse:    0,
				MSpanSys:      0,
				Mallocs:       0,
				NextGC:        0,
				NumForcedGC:   0,
				NumGC:         0,
				OtherSys:      0,
				PauseTotalNs:  0,
				StackInuse:    0,
				StackSys:      0,
				Sys:           0,
				TotalAlloc:    0,
				PollCount:     0,
				RandomValue:   0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewMetricsStorage(), "NewMetricsStorage()")
		})
	}
}
