package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL    = "https://api.open-meteo.com"
	defaultGeoBaseURL = "https://geocoding-api.open-meteo.com"
	defaultCacheTTL   = 30 * time.Minute
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	geoBaseURL string
	cache      *cache
}

type Option func(*Client)

func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

func WithGeoBaseURL(url string) Option {
	return func(c *Client) { c.geoBaseURL = url }
}

func NewClient(httpClient *http.Client, opts ...Option) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	c := &Client{
		httpClient: httpClient,
		baseURL:    defaultBaseURL,
		geoBaseURL: defaultGeoBaseURL,
		cache:      newCache(defaultCacheTTL),
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
		"%s/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m&hourly=temperature_2m,weather_code&daily=weather_code,temperature_2m_max,temperature_2m_min&timezone=auto&forecast_days=7",
		c.baseURL, lat, lon,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return WeatherData{}, fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WeatherData{}, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var raw openMeteoForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return WeatherData{}, fmt.Errorf("decoding weather response: %w", err)
	}

	data := c.transformForecast(raw)
	c.cache.set(lat, lon, data)
	return data, nil
}

func (c *Client) transformForecast(raw openMeteoForecastResponse) WeatherData {
	desc, icon := weatherCodeInfo(raw.Current.WeatherCode)
	current := CurrentWeather{
		Temperature: raw.Current.Temperature,
		FeelsLike:   raw.Current.ApparentTemp,
		Humidity:    raw.Current.Humidity,
		WindSpeed:   raw.Current.WindSpeed,
		WeatherCode: raw.Current.WeatherCode,
		Description: desc,
		Icon:        icon,
	}

	forecast := make([]DailyForecast, len(raw.Daily.Time))
	for i := range raw.Daily.Time {
		d, ic := weatherCodeInfo(raw.Daily.WeatherCode[i])
		forecast[i] = DailyForecast{
			Date:        raw.Daily.Time[i],
			TempMin:     raw.Daily.TempMin[i],
			TempMax:     raw.Daily.TempMax[i],
			WeatherCode: raw.Daily.WeatherCode[i],
			Description: d,
			Icon:        ic,
		}
	}

	now := time.Now()
	var hourly []HourlyForecast
	for i := range raw.Hourly.Time {
		t, err := time.Parse("2006-01-02T15:04", raw.Hourly.Time[i])
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
		hourly = append(hourly, HourlyForecast{
			Time:        raw.Hourly.Time[i],
			Temperature: raw.Hourly.Temperature[i],
			WeatherCode: raw.Hourly.WeatherCode[i],
			Description: d,
			Icon:        ic,
		})
	}

	return WeatherData{Current: current, Hourly: hourly, Forecast: forecast}
}

func (c *Client) Geocode(city string) ([]Location, error) {
	reqURL := fmt.Sprintf("%s/v1/search?name=%s&count=5&language=en", c.geoBaseURL, url.QueryEscape(city))

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode API returned status %d", resp.StatusCode)
	}

	var raw openMeteoGeoResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding geocode response: %w", err)
	}

	locations := make([]Location, len(raw.Results))
	for i, r := range raw.Results {
		locations[i] = Location{
			Name:      r.Name,
			Country:   r.Country,
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
		}
	}
	return locations, nil
}
