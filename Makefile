include .envrc
MIGRATIONS_PATH=./cmd/db/migrations

run:
	go run cmd/api/*.go

curl:
	curl -v http://localhost:8080/v1/health

tidy:
	go mod tidy
	go mod vendor

direnv:
	direnv allow .


migration:
	migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

migrate-down:
	migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down


install:
	brew install direnv
	brew install golang-migrate
	go install github.com/air-verse/air@latest


seed:
	go run cmd/db/seed/main.go


gen-docs:
	swag init -g ./api/main.go -d cmd,internal && swag fmt


add:
	git add -A


.PHONY: run curl tidy direnv migration migrate-up migrate-down install seed gen-docs add
