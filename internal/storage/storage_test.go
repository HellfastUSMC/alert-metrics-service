package storage

import "testing"

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Metrics:   tt.fields.Metrics,
				PollCount: tt.fields.PollCount,
			}
			if err := m.SetMetric(tt.args.metricName, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("SetMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
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
		})
	}
}

func TestMetrics_SendMetrics(t *testing.T) {
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
	type args struct {
		urlAndPort string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
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
			m.SendMetrics(tt.args.urlAndPort)
		})
	}
}
