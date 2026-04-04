package weather

// WeatherData is the response returned by the weather endpoint.
type WeatherData struct {
	Current  CurrentWeather   `json:"current"`
	Hourly   []HourlyForecast `json:"hourly"`
	Forecast []DailyForecast  `json:"forecast"`
}

type CurrentWeather struct {
	Temperature float64 `json:"temperature"`
	FeelsLike   float64 `json:"feelsLike"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"windSpeed"`
	WeatherCode int     `json:"weatherCode"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

type HourlyForecast struct {
	Time        string  `json:"time"`
	Temperature float64 `json:"temperature"`
	WeatherCode int     `json:"weatherCode"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
}

type DailyForecast struct {
	Date        string  `json:"date"`
	TempMin     float64 `json:"tempMin"`
	TempMax     float64 `json:"tempMax"`
	WeatherCode int     `json:"weatherCode"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
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
	Temperature   float64 `json:"temperature_2m"`
	Humidity      int     `json:"relative_humidity_2m"`
	ApparentTemp  float64 `json:"apparent_temperature"`
	WeatherCode   int     `json:"weather_code"`
	WindSpeed     float64 `json:"wind_speed_10m"`
}

type openMeteoHourly struct {
	Time        []string  `json:"time"`
	Temperature []float64 `json:"temperature_2m"`
	WeatherCode []int     `json:"weather_code"`
}

type openMeteoDaily struct {
	Time        []string  `json:"time"`
	WeatherCode []int     `json:"weather_code"`
	TempMax     []float64 `json:"temperature_2m_max"`
	TempMin     []float64 `json:"temperature_2m_min"`
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
