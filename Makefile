run:
	@go run ./cmd/webapp

run/migrate:
	@go run ./cmd/webapp -migrate=true