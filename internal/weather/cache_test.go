package weather

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := newCache(1 * time.Hour)

	data := WeatherData{
		Current: CurrentWeather{Temperature: 20.5},
	}
	c.set(52.52, 13.41, data)

	got, ok := c.get(52.52, 13.41)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Current.Temperature != 20.5 {
		t.Fatalf("expected 20.5, got %f", got.Current.Temperature)
	}
}

func TestCache_Miss(t *testing.T) {
	c := newCache(1 * time.Hour)

	_, ok := c.get(52.52, 13.41)
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCache_Expired(t *testing.T) {
	c := newCache(1 * time.Millisecond)

	c.set(52.52, 13.41, WeatherData{})
	time.Sleep(5 * time.Millisecond)

	_, ok := c.get(52.52, 13.41)
	if ok {
		t.Fatal("expected cache miss after expiry")
	}
}

func TestCache_RoundsCoordinates(t *testing.T) {
	c := newCache(1 * time.Hour)

	c.set(52.5234, 13.4115, WeatherData{
		Current: CurrentWeather{Temperature: 15.0},
	})

	// Slightly different coordinates should hit the same cache entry (both round to 52.52, 13.41)
	got, ok := c.get(52.5241, 13.4143)
	if !ok {
		t.Fatal("expected cache hit for nearby coordinates")
	}
	if got.Current.Temperature != 15.0 {
		t.Fatalf("expected 15.0, got %f", got.Current.Temperature)
	}
}
