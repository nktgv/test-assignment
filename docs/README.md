# HTTP Load Balancer - Документация

## О проекте

HTTP балансировщик нагрузки на Go с поддержкой:
- Round-robin, Least Connections и Random алгоритмов балансировки
- Health-check бэкендов
- Rate-limiting (Token Bucket)
- Конфигурация через YAML файл
- PostgreSQL для хранения состояния

## Требования

- Go 1.21+
- Docker 20.10+
- Docker Compose 2.0+
- golangci-lint (для линтинга)

## Структура проекта

```
.
├── cmd/
│   └── http-load-balancer/
│       └── main.go          # Точка входа
├── config.yaml              # Конфигурация
├── init/                    # Инициализация БД
├── Dockerfile               # Конфигурация Docker
├── docker-compose.yml       # Оркестрация сервисов
├── go.mod                   # Зависимости
└── Taskfile.yml             # Автоматизация задач
```

## Установка и запуск

### 1. Локальный запуск (без Docker)

```bash
# Установите зависимости
go mod download

# Запуск с конфигом по умолчанию
task run

# Или с указанием конфига
go run ./cmd/http-load-balancer/main.go --config=/path/to/config
```

### 2. Запуск через Docker

```bash
# Сборка и запуск всех сервисов
task up

# Остановка
task down

# Полный перезапуск
task restart
```

## Конфигурация

Пример `config.yaml`:

```yaml
env: dev
port: 8090
strategy: round-robin  # или least_connections, random

postgres:
  host: postgres
  port: 5432
  user: lb_user
  password: lb_password
  db: loadbalancer

backends:
  - url: "http://backend1:8080"
    is_alive: true
  - url: "http://backend2:8080"
    is_alive: true

healthcheck:
  interval: 30s
  timeout: 5s

rate_limiting:
  default_capacity: 100
  default_rate: 10
```

## Работа с Taskfile

Основные команды:

```bash
# Запуск линтера
task lint

# Форматирование кода
task lint:format

# Создание бэкапа БД
task backup

# Восстановление БД
task restore backup_file.sql
```

## Рекомендации по разработке

1. **Линтинг**:
   ```bash
   task lint:fix
   ```

## Устранение неполадок

Если контейнер падает:
```bash
docker-compose logs -f loadbalancer
```

Проверить состояние балансера:
```bash
docker-compose exec loadbalancer \
  curl -X GET http://localhost:8090
```