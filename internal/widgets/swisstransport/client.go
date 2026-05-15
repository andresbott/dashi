package swisstransport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL  = "https://transport.opendata.ch"
	defaultCacheTTL = 30 * time.Second
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	cache      *cache
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    defaultBaseURL,
		cache:      newCache(defaultCacheTTL),
	}
}

// SetBaseURL overrides the base URL (for testing).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func (c *Client) SearchStations(query string) ([]Station, error) {
	reqURL := fmt.Sprintf("%s/v1/locations?query=%s&type=station", c.baseURL, url.QueryEscape(query))

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("station search request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("station search API returned status %d", resp.StatusCode)
	}

	var raw locationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding station search response: %w", err)
	}

	stations := make([]Station, 0, len(raw.Stations))
	for _, s := range raw.Stations {
		stations = append(stations, Station{
			ID:        s.ID,
			Name:      s.Name,
			Latitude:  s.Coordinate.X,
			Longitude: s.Coordinate.Y,
		})
	}
	return stations, nil
}

func (c *Client) GetDepartures(stationID string, limit int) ([]Departure, error) {
	if deps, ok := c.cache.get(stationID, limit); ok {
		return deps, nil
	}

	reqURL := fmt.Sprintf("%s/v1/stationboard?id=%s&limit=%d", c.baseURL, url.QueryEscape(stationID), limit)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("stationboard request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stationboard API returned status %d", resp.StatusCode)
	}

	var raw stationboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding stationboard response: %w", err)
	}

	deps := make([]Departure, 0, len(raw.Stationboard))
	for _, entry := range raw.Stationboard {
		scheduled := time.Unix(entry.Stop.DepartureTimestamp, 0)

		expected := scheduled
		if entry.Stop.Prognosis.Departure != nil {
			if t, err := time.Parse("2006-01-02T15:04:05-0700", *entry.Stop.Prognosis.Departure); err == nil {
				expected = t
			}
		}

		deps = append(deps, Departure{
			Category:    entry.Category,
			Number:      entry.Number,
			Destination: entry.To,
			Scheduled:   scheduled,
			Expected:    expected,
			Delay:       entry.Stop.Delay,
			Platform:    entry.Stop.Platform,
		})
	}

	c.cache.set(stationID, limit, deps)
	return deps, nil
}

// API response types

type locationsResponse struct {
	Stations []apiStation `json:"stations"`
}

type apiStation struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Coordinate apiCoordinate `json:"coordinate"`
}

type apiCoordinate struct {
	Type string  `json:"type"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

type stationboardResponse struct {
	Stationboard []apiJourney `json:"stationboard"`
}

type apiJourney struct {
	Category string  `json:"category"`
	Number   string  `json:"number"`
	To       string  `json:"to"`
	Stop     apiStop `json:"stop"`
}

type apiStop struct {
	Departure          string       `json:"departure"`
	DepartureTimestamp int64        `json:"departureTimestamp"`
	Delay              int          `json:"delay"`
	Platform           string       `json:"platform"`
	Prognosis          apiPrognosis `json:"prognosis"`
}

type apiPrognosis struct {
	Departure *string `json:"departure"`
}
