MIGRATIONS_PATH= ./cmd/migrate/migrations

.PHONY : migrate-create
migration:
	@migrate create -seq -ext sql -dir ${MIGRATIONS_PATH} $(filter-out $@,$(MAKECMDGOALS))

.PHONY : migrate-up
migrate-up:
	@migrate --path=${MIGRATIONS_PATH} --database="postgres://admin:adminpassword@localhost:5432/social?sslmode=disable" up

.PHONY : migrate-down
migrate-down:
	@migrate --path=${MIGRATIONS_PATH} --database="postgres://admin:adminpassword@localhost:5432/social?sslmode=disable" down


.PHONY : seed
seed:
	@echo "Seeding the database..."
	@go run ./cmd/migrate/seed/main.go
	@echo "Database seeded successfully."

.PHONY : gen-docs
gen-docs:
	@echo "Generating API documentation..."
	@swag init -g ./api/main.go -d cmd,internal && swag fmt
	@echo "API documentation generated successfully."

.PHONY :  test
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests completed."