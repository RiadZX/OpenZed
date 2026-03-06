package ui

import "testing"

func TestGetOutboundIP(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		//{
		//	name:    "Test GetOutboundIP",
		//	want:    "86.80.172.76",
		//	wantErr: false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOutboundIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOutboundIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetOutboundIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
