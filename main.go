package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Coordinates []struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

type WeatherResponse struct {
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Name string `json:"name"`
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "City parameter is required", http.StatusBadRequest)
		return
	}

	// Get API key from environment variables
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey == "" {
		http.Error(w, "API key not configured", http.StatusInternalServerError)
		return
	}

	// TODO: Get longitude and latitude of city via OpenWeatherMap API
	geoURL := fmt.Sprintf("https://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", city, apiKey)
	geoResp, geoErr := http.Get(geoURL)
	if geoErr != nil {
		http.Error(w, "Failed to fetch geodata", http.StatusInternalServerError)
		return
	}
	defer geoResp.Body.Close()

	var coordinates Coordinates
	if err := json.NewDecoder(geoResp.Body).Decode(&coordinates); err != nil {
		http.Error(w, "Failed to parse geodata", http.StatusInternalServerError)
		return
	}

	if len(coordinates) == 0 {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	// Send weather request with longitude and latitude
	weatherURL := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%s&lon=%s&exclude=minutely,hourly,daily,alerts&appid=%s&units=metric", coordinates[0].Latitude, coordinates[0].Longitude, apiKey)

	weatherResp, weatherErr := http.Get(weatherURL)
	if weatherErr != nil {
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}
	defer weatherResp.Body.Close()

	var weather WeatherResponse
	if err := json.NewDecoder(weatherResp.Body).Decode(&weather); err != nil {
		http.Error(w, "Failed to parse weather data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weather)
}

func main() {
	http.HandleFunc("/weather", weatherHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	port := "8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
