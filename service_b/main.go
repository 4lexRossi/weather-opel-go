package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	traces "go.opentelemetry.io/otel/trace"
)

var tracer traces.Tracer

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	// Configure Zipkin exporter
	const zipkinEndpoint = "http://zipkin:9411/api/v2/spans"
	exporter, err := zipkin.New(zipkinEndpoint)
	if err != nil {
		log.Fatalf("Failed to configure Zipkin exporter: %v", err)
	}

	// Set up TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes("service-b",
			attribute.String("service.name", "servico-b"),
		)),
	)
	otel.SetTracerProvider(tp)

	tracer = tp.Tracer("servico-b-tracing")

	// Set up HTTP server
	http.HandleFunc("/weather", handleWeatherRequest)

	log.Println("Service B started on port 8082...")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleWeatherRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	cep, exists := req["cep"]
	if !exists {
		http.Error(w, "CEP is required", http.StatusBadRequest)
		return
	}

	ctx, span := tracer.Start(r.Context(), "Fetching weather for CEP")
	defer span.End()

	weatherInfo, err := getWeatherInfo(ctx, cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(weatherInfo); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func getWeatherInfo(ctx context.Context, cep string) (*WeatherResponse, error) {
	ctx, span := tracer.Start(ctx, "Querying city and weather")
	defer span.End()

	// Simulated logic for weather data
	switch cep {
	case "29902555":
		return &WeatherResponse{
			City:  "São Paulo",
			TempC: 28.5,
			TempF: 28.5*1.8 + 32,
			TempK: 28.5 + 273.15,
		}, nil
	default:
		return nil, fmt.Errorf("ZIP code not found")
	}
}
