package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func FetchCityFromViaCEP(ctx context.Context, cep string) (string, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	city, ok := result["localidade"].(string)
	if !ok {
		return "", fmt.Errorf("city not found")
	}
	return city, nil
}

func FetchTemperatureFromWeatherAPI(ctx context.Context, city string) (float64, error) {
	// Replace with WeatherAPI implementation
	return 28.5, nil // Mock value for testing
}
