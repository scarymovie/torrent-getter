DOCKER_COMPOSE = docker compose -f docker/docker-compose.development.yml
DOCKER_EXEC_MIGRATE = docker exec -it torrent-getter-migrate
DB_URL = postgres://db_user:db_password@postgres:5432/db_database?sslmode=disable

# Для первого старта
.PHONE: init
init: up migrate-up

# Запуск контейнеров
.PHONY: up
up:
	$(DOCKER_COMPOSE) up -d --build

# Остановка контейнеров
.PHONY: down
down:
	$(DOCKER_COMPOSE) down

# Применение миграций
.PHONY: migrate-up
migrate-up:
	$(DOCKER_EXEC_MIGRATE) migrate -path=/migrations -database=$(DB_URL) up

# Откат миграций
.PHONY: migrate-down
migrate-down:
	$(DOCKER_EXEC_MIGRATE) migrate -path=/migrations -database=$(DB_URL) down 1

# Полный откат всех миграций
.PHONY: migrate-reset
migrate-reset:
	$(DOCKER_EXEC_MIGRATE) migrate -path=/migrations -database=$(DB_URL) down -all

# Создание новой миграции
.PHONY: migrate-create
migrate-create:
ifndef NAME
	$(error NAME is not set. Usage: make migrate-create NAME=create_users_table)
endif
	$(DOCKER_EXEC_MIGRATE) migrate create -ext sql -dir /migrations -seq $(NAME)

# Просмотр версии миграций
.PHONY: migrate-version
migrate-version:
	$(DOCKER_EXEC_MIGRATE) migrate -path=/migrations -database=$(DB_URL) version