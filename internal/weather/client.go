package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL           = "https://api.open-meteo.com"
	defaultGeoBaseURL        = "https://geocoding-api.open-meteo.com"
	defaultAirQualityBaseURL = "https://air-quality-api.open-meteo.com"
	defaultCacheTTL          = 30 * time.Minute
)

type Client struct {
	httpClient        *http.Client
	baseURL           string
	geoBaseURL        string
	airQualityBaseURL string
	cache             *cache
}

type Option func(*Client)

func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

func WithGeoBaseURL(url string) Option {
	return func(c *Client) { c.geoBaseURL = url }
}

func WithAirQualityBaseURL(url string) Option {
	return func(c *Client) { c.airQualityBaseURL = url }
}

func NewClient(httpClient *http.Client, opts ...Option) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	c := &Client{
		httpClient:        httpClient,
		baseURL:           defaultBaseURL,
		geoBaseURL:        defaultGeoBaseURL,
		airQualityBaseURL: defaultAirQualityBaseURL,
		cache:             newCache(defaultCacheTTL),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) GetWeather(lat, lon float64) (WeatherData, error) {
	if data, ok := c.cache.get(lat, lon); ok {
		return data, nil
	}

	url := fmt.Sprintf(
		"%s/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m,surface_pressure&hourly=temperature_2m,weather_code,visibility,precipitation_probability&daily=weather_code,temperature_2m_max,temperature_2m_min,sunrise,sunset,uv_index_max&timezone=auto&forecast_days=7",
		c.baseURL, lat, lon,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return WeatherData{}, fmt.Errorf("weather API request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return WeatherData{}, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var raw openMeteoForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return WeatherData{}, fmt.Errorf("decoding weather response: %w", err)
	}

	data := c.transformForecast(raw)
	data.AirQuality = c.fetchAirQuality(lat, lon)
	c.cache.set(lat, lon, data)
	return data, nil
}

func (c *Client) fetchAirQuality(lat, lon float64) *AirQuality {
	url := fmt.Sprintf(
		"%s/v1/air-quality?latitude=%f&longitude=%f&current=european_aqi",
		c.airQualityBaseURL, lat, lon,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var raw openMeteoAirQualityResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil
	}

	return &AirQuality{EuropeanAQI: raw.Current.EuropeanAQI}
}

func (c *Client) transformForecast(raw openMeteoForecastResponse) WeatherData {
	desc, icon := weatherCodeInfo(raw.Current.WeatherCode)
	current := CurrentWeather{
		Temperature: raw.Current.Temperature,
		FeelsLike:   raw.Current.ApparentTemp,
		Humidity:    raw.Current.Humidity,
		WindSpeed:   raw.Current.WindSpeed,
		Pressure:    raw.Current.Pressure,
		WeatherCode: raw.Current.WeatherCode,
		Description: desc,
		Icon:        icon,
	}

	// API returns times in the location's local timezone (timezone=auto).
	// Parse them in the server's local timezone so Before() comparisons work.
	now := time.Now()
	loc := now.Location()
	for i := range raw.Hourly.Time {
		t, err := time.ParseInLocation("2006-01-02T15:04", raw.Hourly.Time[i], loc)
		if err != nil {
			continue
		}
		if t.Before(now) {
			continue
		}
		if i < len(raw.Hourly.Visibility) {
			current.Visibility = raw.Hourly.Visibility[i] / 1000.0
		}
		break
	}

	forecast := make([]DailyForecast, len(raw.Daily.Time))
	for i := range raw.Daily.Time {
		d, ic := weatherCodeInfo(raw.Daily.WeatherCode[i])
		f := DailyForecast{
			Date:        raw.Daily.Time[i],
			TempMin:     raw.Daily.TempMin[i],
			TempMax:     raw.Daily.TempMax[i],
			WeatherCode: raw.Daily.WeatherCode[i],
			Description: d,
			Icon:        ic,
		}
		if i < len(raw.Daily.Sunrise) {
			f.Sunrise = raw.Daily.Sunrise[i]
		}
		if i < len(raw.Daily.Sunset) {
			f.Sunset = raw.Daily.Sunset[i]
		}
		if i < len(raw.Daily.UVIndexMax) {
			f.UVIndex = raw.Daily.UVIndexMax[i]
		}
		forecast[i] = f
	}

	var hourly []HourlyForecast
	for i := range raw.Hourly.Time {
		t, err := time.ParseInLocation("2006-01-02T15:04", raw.Hourly.Time[i], loc)
		if err != nil {
			continue
		}
		if t.Before(now) {
			continue
		}
		if len(hourly) >= 24 {
			break
		}
		d, ic := weatherCodeInfo(raw.Hourly.WeatherCode[i])
		hf := HourlyForecast{
			Time:        raw.Hourly.Time[i],
			Temperature: raw.Hourly.Temperature[i],
			WeatherCode: raw.Hourly.WeatherCode[i],
			Description: d,
			Icon:        ic,
		}
		if i < len(raw.Hourly.PrecipitationProbability) {
			hf.PrecipitationProbability = raw.Hourly.PrecipitationProbability[i]
		}
		hourly = append(hourly, hf)
	}

	return WeatherData{Current: current, Hourly: hourly, Forecast: forecast}
}

// WarmupLocations pre-fetches weather data for the given lat/lon pairs,
// populating the cache so subsequent requests are instant.
// It respects context cancellation between iterations.
func (c *Client) WarmupLocations(ctx context.Context, locations [][2]float64) {
	for _, loc := range locations {
		if ctx.Err() != nil {
			return
		}
		// GetWeather checks the cache first, so duplicates are cheap
		_, _ = c.GetWeather(loc[0], loc[1])
	}
}

func (c *Client) Geocode(city string) ([]Location, error) {
	reqURL := fmt.Sprintf("%s/v1/search?name=%s&count=5&language=en", c.geoBaseURL, url.QueryEscape(city))

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("geocode request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode API returned status %d", resp.StatusCode)
	}

	var raw openMeteoGeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding geocode response: %w", err)
	}

	locations := make([]Location, 0, len(raw.Results))
	for _, r := range raw.Results {
		locations = append(locations, Location(r))
	}
	return locations, nil
}
