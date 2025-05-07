package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type WeatherResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Hourly    struct {
		Time                     []string  `json:"time"`
		Temperature2m            []float64 `json:"temperature_2m"`
		PrecipitationProbability []float64 `json:"precipitation_probability"`
		Precipitation            []float64 `json:"precipitation"`
	} `json:"hourly"`
	Daily struct {
		Time                        []string  `json:"time"`
		Temperature2mMax            []float64 `json:"temperature_2m_max"`
		Temperature2mMin            []float64 `json:"temperature_2m_min"`
		PrecipitationSum            []float64 `json:"precipitation_sum"`
		RainSum                     []float64 `json:"rain_sum"`
		PrecipitationHours          []float64 `json:"precipitation_hours"`
		PrecipitationProbabilityMax []float64 `json:"precipitation_probability_max"`
		WindSpeed10mMax             []float64 `json:"wind_speed_10m_max"`
	} `json:"daily"`
}

func GetWeatherForecast(latitude float64, longitude float64) (*WeatherResponse, error) {
	baseURL := "https://api.open-meteo.com/v1/forecast"

	params := url.Values{}
	params.Add("latitude", strconv.FormatFloat(latitude, 'f', -1, 64))
	params.Add("longitude", strconv.FormatFloat(longitude, 'f', -1, 64))
	params.Add("hourly", "temperature_2m,precipitation_probability,precipitation")
	params.Add("daily", "temperature_2m_max,temperature_2m_min,precipitation_sum,rain_sum,precipitation_hours,precipitation_probability_max,wind_speed_10m_max")
	params.Add("timezone", "auto")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	client := &http.Client{}

	resp, err := client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the JSON response
	var weatherResponse WeatherResponse
	if err := json.Unmarshal(body, &weatherResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %w", err)
	}

	return &weatherResponse, nil
}
func main() {
	defaultLat := 40.71 //New York City
	defaultLon := -74.01
	defaultDays := 2

	// Set up command line flags
	latitude := flag.Float64("lat", defaultLat, "Latitude (default: New York City)")
	longitude := flag.Float64("lon", defaultLon, "Longitude (default: New York City)")
	days := flag.Int("days", defaultDays, "Number of days to show (default: 2; max: 7)")
	flag.Parse()

	// Print usage information if requested
	if flag.NFlag() == 0 {
		fmt.Printf("Using default location: New York City (%.2f, %.2f) and %d days\n",
			defaultLat, defaultLon, defaultDays)
		fmt.Println("You can specify location and days with: -lat=<value> -lon=<value> -days=<value>")
	}

	if *days < 1 {
		fmt.Println("Error: Days must be at least 1")
		os.Exit(1)
	}

	response, err := GetWeatherForecast(*latitude, *longitude)
	if err != nil {
		fmt.Printf("Error getting weather forecast: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Weather for: %.4f, %.4f - Timezone: %s\n", response.Latitude, response.Longitude, response.Timezone)

	// Print daily forecast for specified number of days
	daysToShow := *days
	if len(response.Daily.Time) < daysToShow {
		daysToShow = len(response.Daily.Time)
	}

	for i := 0; i < daysToShow; i++ {
		var dayLabel string
		if i == 0 {
			dayLabel = "Today"
		} else if i == 1 {
			dayLabel = "Tomorrow"
		} else {
			dayLabel = fmt.Sprintf("Day %d", i+1)
		}

		fmt.Printf("%s (%s):\n", dayLabel, response.Daily.Time[i])
		fmt.Printf("  Temperature: %.1f°C to %.1f°C\n",
			response.Daily.Temperature2mMin[i],
			response.Daily.Temperature2mMax[i])
		fmt.Printf("  Precipitation: %.1f mm (probability: %.1f%%)\n",
			response.Daily.PrecipitationSum[i],
			response.Daily.PrecipitationProbabilityMax[i])
		fmt.Printf("  Rain: %.1f mm - Precipitation Hours: %.1f\n", response.Daily.RainSum[i],
			response.Daily.PrecipitationHours[i])
		fmt.Printf("  Max Wind Speed: %.1f km/h\n\n", response.Daily.WindSpeed10mMax[i])
	}

	hoursToPrint := 5
	fmt.Printf("Hourly Forecast (next %d hours):\n", hoursToPrint)
	if len(response.Hourly.Time) < hoursToPrint {
		hoursToPrint = len(response.Hourly.Time)
	}

	for j := 0; j < hoursToPrint; j++ {
		fmt.Printf("  %s: %.1f°C, Precipitation: %.1f mm (%.1f%% probability)\n",
			response.Hourly.Time[j],
			response.Hourly.Temperature2m[j],
			response.Hourly.Precipitation[j],
			response.Hourly.PrecipitationProbability[j])
	}
}
