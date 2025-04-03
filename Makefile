.PHONY: dev-backend dev-frontend dev build clean

# Development commands
dev-backend:
	docker-compose -f docker-compose.dev.yml up -d

dev-frontend:
	cd frontend && npm start

# Run both in separate terminals (use this in documentation)
dev:
	@echo "Run these commands in separate terminals:"
	@echo "make dev-backend"
	@echo "make dev-frontend"

# Build commands
build:
	docker-compose build --no-cache

# Clean up
clean:
	docker-compose down --volumes --remove-orphans
	docker-compose -f docker-compose.dev.yml down --volumes --remove-orphans

# Help
help:
	@echo "Available commands:"
	@echo "  make dev-backend    - Start the backend in development mode"
	@echo "  make dev-frontend   - Start the frontend in development mode"
	@echo "  make dev            - Instructions to start both frontend and backend"
	@echo "  make build          - Build the production Docker images"
	@echo "  make clean          - Clean up all containers and volumes" 