.PHONY: help build up down logs clean test

help:
	@echo "Available commands:"
	@echo "  make build       - Build Docker images"
	@echo "  make up          - Start services with docker-compose"
	@echo "  make down        - Stop services"
	@echo "  make logs        - View logs"
	@echo "  make clean       - Remove containers and volumes"
	@echo "  make test        - Run tests"

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

clean:
	docker-compose down -v

test:
	cd backend && go test ./...

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

dev: build up logs
