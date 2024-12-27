run:
	go run cmd/api/*.go

curl:
	curl http://localhost:8080/v1/health

tidy:
	go mod tidy
	go mod vendor


