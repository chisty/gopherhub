run:
	go run cmd/api/*.go

curl:
	curl http://localhost:8080/v1/health

tidy:
	go mod tidy
	go mod vendor

direnv:
	direnv allow .

db-up:
	migrate -path=./cmd/db/migrations -database "postgres://postgres:postgres@localhost:5432/gopherhub?sslmode=disable" up

db-down:
	migrate -path=./cmd/db/migrations -database "postgres://postgres:postgres@localhost:5432/gopherhub?sslmode=disable" down


install:
	brew install direnv
	brew install golang-migrate
	go install github.com/air-verse/air@latest
