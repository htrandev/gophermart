test:
	@go test -short -race -timeout 30s -coverprofile=cover.out ./... 

cover:
	@go tool cover -func=cover.out  

migration.sql:
	@goose -dir ./deployments/migrations/postgres create $(name) sql 