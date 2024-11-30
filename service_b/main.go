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
	// Update Zipkin endpoint to explicitly use IPv4
	zipkinEndpoint := "http://127.0.0.1:9411/api/v2/spans"
	exporter, err := zipkin.New(zipkinEndpoint)
	if err != nil {
		log.Fatalf("Failed to configure Zipkin exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes("service-b", attribute.String("service.name", "servico-b"))),
	)
	otel.SetTracerProvider(tp)

	tracer = tp.Tracer("servico-b-tracing")

	http.HandleFunc("/weather", handleWeatherRequest)

	log.Println("Servidor B iniciado na porta 8082...")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func handleWeatherRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Erro ao ler corpo da requisição", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Formato inválido", http.StatusBadRequest)
		return
	}

	cep := req["cep"]

	ctx, span := tracer.Start(r.Context(), "Consultando clima para o CEP")
	defer span.End()

	weatherInfo, err := getWeatherInfo(ctx, cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(weatherInfo)
}

func getWeatherInfo(ctx context.Context, cep string) (*WeatherResponse, error) {
	ctx, span := tracer.Start(ctx, "Consultando a cidade e clima")
	defer span.End()

	if cep == "29902555" {
		return &WeatherResponse{
			City:  "São Paulo",
			TempC: 28.5,
			TempF: 28.5*1.8 + 32,
			TempK: 28.5 + 273,
		}, nil
	}

	return nil, fmt.Errorf("can not find zipcode")
}
