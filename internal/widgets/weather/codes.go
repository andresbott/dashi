package weather

type codeInfo struct {
	description string
	icon        string
}

var weatherCodes = map[int]codeInfo{
	0:  {"Clear sky", "clear-sky"},
	1:  {"Mainly clear", "mainly-clear"},
	2:  {"Partly cloudy", "partly-cloudy"},
	3:  {"Overcast", "overcast"},
	45: {"Foggy", "foggy"},
	48: {"Depositing rime fog", "foggy"},
	51: {"Light drizzle", "drizzle-light"},
	53: {"Moderate drizzle", "drizzle-moderate"},
	55: {"Dense drizzle", "drizzle-dense"},
	56: {"Light freezing drizzle", "freezing-drizzle-light"},
	57: {"Dense freezing drizzle", "freezing-drizzle-dense"},
	61: {"Slight rain", "rain-slight"},
	63: {"Moderate rain", "rain-moderate"},
	65: {"Heavy rain", "rain-heavy"},
	66: {"Light freezing rain", "freezing-rain-light"},
	67: {"Heavy freezing rain", "freezing-rain-heavy"},
	71: {"Slight snowfall", "snow-slight"},
	73: {"Moderate snowfall", "snow-moderate"},
	75: {"Heavy snowfall", "snow-heavy"},
	77: {"Snow grains", "snow-grains"},
	80: {"Slight rain showers", "rain-showers-slight"},
	81: {"Moderate rain showers", "rain-showers-moderate"},
	82: {"Violent rain showers", "rain-showers-violent"},
	85: {"Slight snow showers", "snow-showers-slight"},
	86: {"Heavy snow showers", "snow-showers-heavy"},
	95: {"Thunderstorm", "thunderstorm"},
	96: {"Thunderstorm with slight hail", "thunderstorm-hail-slight"},
	99: {"Thunderstorm with heavy hail", "thunderstorm-hail-heavy"},
}

func weatherCodeInfo(code int) (description string, icon string) {
	if info, ok := weatherCodes[code]; ok {
		return info.description, info.icon
	}
	return "Unknown", "unknown"
}
