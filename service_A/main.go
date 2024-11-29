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
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

type CepRequest struct {
	Cep string `json:"cep"`
}

func main() {
	// Configuração do OpenTelemetry
	exporter, err := zipkin.NewExporter("http://localhost:9411/api/v2/spans", zipkin.WithSDKOptions())
	if err != nil {
		log.Fatalf("Falha ao configurar o exportador Zipkin: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(attribute.Key("service.name").String("servico-a"))),
	)
	otel.SetTracerProvider(tp)

	// Obter o tracer
	tracer = tp.Tracer("servico-a-tracing")

	http.HandleFunc("/cep", handleCepRequest)

	log.Println("Servidor A iniciado na porta 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
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

	// Valida o CEP
	if !isValidCEP(req.Cep) {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// Criar span para rastrear a chamada ao Serviço B
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
	// Criar span para a chamada ao Serviço B
	ctx, span := tracer.Start(ctx, "Chamada ao Serviço B")
	defer span.End()

	// Envia a requisição HTTP para o Serviço B
	requestBody, err := json.Marshal(map[string]string{"cep": cep})
	if err != nil {
		return nil, fmt.Errorf("erro ao montar o corpo da requisição para o Serviço B")
	}

	resp, err := http.Post("http://localhost:8082/weather", "application/json", bytes.NewBuffer(requestBody))
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
