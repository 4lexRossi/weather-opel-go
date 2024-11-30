package main

import (
	"log"
	"net/http"

	"github.com/4lexRossi/service-b-weather-opel-go/utils"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

type WeatherRequest struct {
	CEP string `json:"cep"`
}

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func main() {
	// Initialize OTEL tracing
	tp := utils.InitTracer("service-b")

	// Initialize Fiber app
	app := fiber.New()

	tracer = tp.Tracer("service-b-tracing")

	// Define routes
	app.Post("/weather", handleWeatherRequest)

	// Start the server
	log.Println("Service B running on http://localhost:8082")
	log.Fatal(app.Listen(":8082"))
}

func handleWeatherRequest(c *fiber.Ctx) error {
	// Start tracing span
	ctx, span := tracer.Start(c.Context(), "handleWeatherRequest")
	defer span.End()

	// Parse request body
	var req WeatherRequest
	if err := c.BodyParser(&req); err != nil || !isValidPostcode(req.CEP) {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"message": "invalid zipcode"})
	}

	// Fetch location using ViaCEP
	city, err := utils.FetchCityFromViaCEP(ctx, req.CEP)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"message": "can not find zipcode"})
	}

	// Fetch temperature from WeatherAPI
	tempC, err := utils.FetchTemperatureFromWeatherAPI(ctx, city)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "error fetching temperature"})
	}

	// Convert temperatures
	tempF := utils.CelsiusToFahrenheit(tempC)
	tempK := utils.CelsiusToKelvin(tempC)

	// Respond with the result
	response := WeatherResponse{
		City:  city,
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}
	return c.Status(http.StatusOK).JSON(response)
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
