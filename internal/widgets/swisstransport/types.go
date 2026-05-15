package swisstransport

import "time"

type Station struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Departure struct {
	Category    string    `json:"category"`
	Number      string    `json:"number"`
	Destination string    `json:"destination"`
	Scheduled   time.Time `json:"scheduled"`
	Expected    time.Time `json:"expected"`
	Delay       int       `json:"delay"`
	Platform    string    `json:"platform"`
}
