package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	tracers "go.opentelemetry.io/otel/trace"
)

var tracer tracers.Tracer

type CepRequest struct {
	Cep string `json:"cep"`
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
		trace.WithResource(resource.NewWithAttributes("service-a", attribute.String("service.name", "servico-a"))),
	)
	otel.SetTracerProvider(tp)

	tracer = tp.Tracer("servico-a-tracing")

	http.HandleFunc("/cep", handleCepRequest)

	log.Println("Servidor A iniciado na porta 8081...")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleCepRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req CepRequest
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

	if !isValidCEP(req.Cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	ctx, span := tracer.Start(r.Context(), "Chamando o Serviço B")
	defer span.End()

	cityInfo, err := callServicoB(ctx, req.Cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cityInfo)
}

func isValidCEP(cep string) bool {
	re := regexp.MustCompile(`^\d{8}$`)
	return re.MatchString(cep)
}

func callServicoB(ctx context.Context, cep string) (map[string]interface{}, error) {
	ctx, span := tracer.Start(ctx, "Chamada ao Serviço B")
	defer span.End()

	requestBody, err := json.Marshal(map[string]string{"cep": cep})
	if err != nil {
		return nil, fmt.Errorf("erro ao montar o corpo da requisição para o Serviço B")
	}

	resp, err := http.Post("http://127.0.0.1:8082/weather", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar o Serviço B: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Serviço B retornou erro: %s", string(body))
	}

	var cityInfo map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta do Serviço B")
	}

	err = json.Unmarshal(body, &cityInfo)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta do Serviço B")
	}

	return cityInfo, nil
}
