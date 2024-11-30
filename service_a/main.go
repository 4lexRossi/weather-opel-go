package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/4lexRossi/service-a-weather-opel-go/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

type PostcodeRequest struct {
	CEP string `json:"cep"`
}

func main() {
	// Initialize OTEL tracing
	tp := utils.InitTracer("service-a")

	// Initialize Fiber app
	app := fiber.New()

	tracer = tp.Tracer("service-a-tracing")

	// Define routes
	app.Post("/cep", validatePostcode)

	// Start the server
	log.Println("Service A running on http://localhost:8081")
	log.Fatal(app.Listen(":8081"))
}

func validatePostcode(c *fiber.Ctx) error {
	// Start tracing span
	ctx, span := tracer.Start(c.Context(), "validatePostcode")
	defer span.End()

	// Parse request body
	var req PostcodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "invalid zipcode"})
	}

	// Validate postcode
	if !isValidPostcode(req.CEP) {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "invalid zipcode"})
	}

	// Forward to Service B
	forwardedResponse, err := forwardToServiceB(ctx, req.CEP)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "error forwarding request"})
	}
	defer forwardedResponse.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(forwardedResponse.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "error reading response"})
	}

	var cityInfo map[string]interface{}
	err = json.Unmarshal(bodyBytes, &cityInfo)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON("erro ao decodificar resposta do Serviço B")
	}

	// Respond with Service B's response
	return c.Status(forwardedResponse.StatusCode).JSON(cityInfo)
}

func forwardToServiceB(ctx context.Context, cep string) (*http.Response, error) {
	client := retryablehttp.NewClient()
	client.RetryMax = 3

	requestBody, _ := json.Marshal(PostcodeRequest{CEP: cep})
	req, err := retryablehttp.NewRequest("POST", "http://localhost:8082/weather", requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func isValidPostcode(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for _, char := range cep {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
