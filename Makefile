
build:
	cd backend && go build -o air-compose ./cmd/air-compose

run:
	cd backend && go run ./cmd/air-compose run

run-dev:
	ENV=dev cd backend && go run ./cmd/air-compose run

up-local:
	docker compose -f compose.local.yaml up --build

lint:
	cd backend && golangci-lint run ./...

test:
	cd backend && go test ./internal/...

test-int:
	(docker stop air-compose || true) && \
	(docker rm air-compose || true) && \
	(docker rmi air-compose:local --force || true) && \
	cd backend && go test ./integration_tests/... -v -count=1

test-cover:
	cd backend && go test -coverprofile=coverage.out ./internal/... && \
	go tool cover -html=coverage.out

gen-api : 
	make tsp-gen 
	make oapi-gen
	make orval-gen

tsp-gen:
	cd api && npm run tsp

oapi-gen:
	cd backend && go tool oapi-codegen -config oapi-codegen.yaml ../api/tsp-output/schema/openapi.1.0.yaml

orval-gen:
	cd frontend && npm run orval

# docker-run:
# 	docker compose --env-file ./_ignore_.env up --build