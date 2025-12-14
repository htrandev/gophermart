
test:
	@go test -short -race -timeout 30s -coverprofile=cover.out ./... 

cover:
	@go tool cover -func=cover.out  

migration.sql:
	@goose -dir ./deployments/migrations/postgres create $(name) sql 

.PHONY: request.gen.easyjson 
request.gen.easyjson:
	@easyjson -all -output_filename=internal/application/request_easyjson.gen.go internal/application/request.go 

.PHONY: response.gen.easyjson 
response.gen.easyjson:
	@easyjson -all -output_filename=internal/application/response_easyjson.gen.go internal/application/response.go 

.PHONY: domain.accural.gen.easyjson 
domain.accural.gen.easyjson:
	@easyjson -all -output_filename=internal/domain/accural_easyjson.gen.go internal/domain/accural.go 

.PHONY: domain.billing.gen.easyjson 
domain.billing.gen.easyjson:
	@easyjson -all -output_filename=internal/domain/billing_easyjson.gen.go internal/domain/billing.go 

.PHONY: domain.order.gen.easyjson 
domain.order.gen.easyjson:
	@easyjson -all -output_filename=internal/domain/order_easyjson.gen.go internal/domain/order.go 

.PHONY: domain.user.gen.easyjson 
domain.user.gen.easyjson:
	@easyjson -all -output_filename=internal/domain/user_easyjson.gen.go internal/domain/user.go 

.PHONY: gen.all.easyjson
gen.all.easyjson: request.gen.easyjson \
	response.gen.easyjson \
	domain.accural.gen.easyjson \
	domain.billing.gen.easyjson \
	domain.order.gen.easyjson \
	domain.user.gen.easyjson

