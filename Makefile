# Variables
DOCKER_COMPOSE_FILE=docker-compose.yaml

.PHONY: up
up:
	@echo "Starting all services with Docker Compose..."
	docker compose -f $(DOCKER_COMPOSE_FILE) up --build

# Stop services with Docker Compose
.PHONY: down
down:
	@echo "Stopping all services..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down

# Clean Docker containers and images
.PHONY: clean
clean:
	@echo "Cleaning Docker containers and images..."
	docker compose -f $(DOCKER_COMPOSE_FILE) down -v --rmi all
	docker system prune -f
