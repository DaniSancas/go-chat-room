# Makefile

all_configs = docker compose -f ./docker-compose-server.yml -f ./docker-compose-client.yml

up:
	$(all_configs) up -d --remove-orphans

down:
	$(all_configs) down --remove-orphans

logs:
	$(all_configs) logs -f --timestamps

build:
	(cd ./server && go build -o ./bin/server ./cmd/api/main.go)

test:
	./scripts/test_all.sh ./server