package weather

// WeatherData is the response returned by the weather endpoint.
type WeatherData struct {
	Current    CurrentWeather   `json:"current"`
	Hourly     []HourlyForecast `json:"hourly"`
	Forecast   []DailyForecast  `json:"forecast"`
	AirQuality *AirQuality      `json:"airQuality,omitempty"`
}

type CurrentWeather struct {
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feelsLike"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"windSpeed"`
	Pressure    float64 `json:"pressure"`
	Visibility  float64 `json:"visibility"`
	WeatherCode int     `json:"weatherCode"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

type HourlyForecast struct {
	Time                     string  `json:"time"`
	Temperature              float64 `json:"temperature"`
	PrecipitationProbability float64 `json:"precipitationProbability"`
	WeatherCode              int     `json:"weatherCode"`
	Description              string  `json:"description"`
	Icon                     string  `json:"icon"`
}

type DailyForecast struct {
	Date        string  `json:"date"`
	TempMin     float64 `json:"tempMin"`
	TempMax     float64 `json:"tempMax"`
	Sunrise     string  `json:"sunrise"`
	Sunset      string  `json:"sunset"`
	UVIndex     float64 `json:"uvIndex"`
	WeatherCode int     `json:"weatherCode"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

type AirQuality struct {
	EuropeanAQI int `json:"europeanAqi"`
}

// Location is a geocoding result.
type Location struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Open-Meteo API response types (internal, for JSON unmarshalling)

type openMeteoForecastResponse struct {
	Current openMeteoCurrent `json:"current"`
	Hourly  openMeteoHourly  `json:"hourly"`
	Daily   openMeteoDaily   `json:"daily"`
}

type openMeteoCurrent struct {
	Temperature  float64 `json:"temperature_2m"`
	Humidity     int     `json:"relative_humidity_2m"`
	ApparentTemp float64 `json:"apparent_temperature"`
	WeatherCode  int     `json:"weather_code"`
	WindSpeed    float64 `json:"wind_speed_10m"`
	Pressure     float64 `json:"surface_pressure"`
}

type openMeteoHourly struct {
	Time                     []string  `json:"time"`
	Temperature              []float64 `json:"temperature_2m"`
	WeatherCode              []int     `json:"weather_code"`
	Visibility               []float64 `json:"visibility"`
	PrecipitationProbability []float64 `json:"precipitation_probability"`
}

type openMeteoDaily struct {
	Time        []string  `json:"time"`
	WeatherCode []int     `json:"weather_code"`
	TempMax     []float64 `json:"temperature_2m_max"`
	TempMin     []float64 `json:"temperature_2m_min"`
	Sunrise     []string  `json:"sunrise"`
	Sunset      []string  `json:"sunset"`
	UVIndexMax  []float64 `json:"uv_index_max"`
}

type openMeteoGeoResponse struct {
	Results []openMeteoGeoResult `json:"results"`
}

type openMeteoGeoResult struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type openMeteoAirQualityResponse struct {
	Current openMeteoAirQualityCurrent `json:"current"`
}

type openMeteoAirQualityCurrent struct {
	EuropeanAQI int `json:"european_aqi"`
}
