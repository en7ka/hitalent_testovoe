.PHONY: run test tidy docker-up docker-down migrate-up

run:
	go run ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

docker-up:
	docker compose up --build

docker-down:
	docker compose down

migrate-up:
	goose -dir ./migrations postgres "$$DATABASE_URL" up
