package weather

import "testing"

func TestWeatherCodeInfo(t *testing.T) {
	tests := []struct {
		code     int
		wantDesc string
		wantIcon string
	}{
		{0, "Clear sky", "clear-sky"},
		{1, "Mainly clear", "mainly-clear"},
		{2, "Partly cloudy", "partly-cloudy"},
		{3, "Overcast", "overcast"},
		{45, "Foggy", "foggy"},
		{51, "Light drizzle", "drizzle-light"},
		{61, "Slight rain", "rain-slight"},
		{71, "Slight snowfall", "snow-slight"},
		{95, "Thunderstorm", "thunderstorm"},
		{999, "Unknown", "unknown"},
	}
	for _, tt := range tests {
		desc, icon := weatherCodeInfo(tt.code)
		if desc != tt.wantDesc {
			t.Errorf("weatherCodeInfo(%d) desc = %q, want %q", tt.code, desc, tt.wantDesc)
		}
		if icon != tt.wantIcon {
			t.Errorf("weatherCodeInfo(%d) icon = %q, want %q", tt.code, icon, tt.wantIcon)
		}
	}
}
