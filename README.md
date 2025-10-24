# IronNode - Blockchain Node Infrastructure Platform

Платформа для доступа к blockchain нодам через API, построенная на микросервисной архитектуре с использованием Go, gRPC и REST API.

## Архитектура

Проект состоит из следующих микросервисов:

### 1. API Gateway (`:8080`)
- REST API endpoint для клиентских запросов
- Маршрутизация запросов к другим сервисам
- Rate limiting и CORS middleware
- Аутентификация JWT токенов

### 2. Auth Service (`:50051` - gRPC)
- Регистрация и аутентификация пользователей
- Генерация и валидация JWT токенов
- Управление пользовательскими сессиями
- Хеширование паролей с bcrypt

### 3. User Service (`:50052` - gRPC)
- Управление API ключами пользователей
- CRUD операции для API ключей
- Валидация API ключей

### 4. Blockchain Service (`:50053` - gRPC)
- Управление подключениями к blockchain нодам
- Поддержка множественных блокчейнов (Ethereum, Polygon, BSC, и др.)
- Приоритизация нод

### 5. Analytics Service (`:50055` - gRPC)
- Логирование всех запросов
- Статистика использования
- История запросов пользователей
- Метрики производительности

### 6. Billing Service (`:50056` - gRPC)
- Управление подписками (Free, Basic, Professional, Enterprise)
- Отслеживание использования квот
- Проверка лимитов запросов

## Технологический стек

- **Backend**: Go 1.21
- **Web Framework**: Gin
- **RPC**: gRPC + Protocol Buffers
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: RabbitMQ (для асинхронных задач)
- **ORM**: GORM
- **Authentication**: JWT
- **Containerization**: Docker & Docker Compose
- **Concurrency**: Goroutines + Channels (Async Logger, Worker Pool, Parallel Requests)

## Структура проекта

\`\`\`
quicknode-clone/
├── services/
│   ├── api-gateway/           # HTTP API Gateway
│   ├── auth-service/          # Аутентификация
│   ├── user-service/          # Управление пользователями и API ключами
│   ├── blockchain-service/    # Управление blockchain нодами
│   ├── analytics-service/     # Аналитика и логирование
│   └── billing-service/       # Биллинг и подписки
├── pkg/
│   ├── config/               # Конфигурация
│   ├── database/             # Подключение к БД
│   ├── cache/                # Redis клиент
│   ├── logger/               # Логирование
│   ├── middleware/           # HTTP middleware
│   ├── models/               # Database модели
│   └── response/             # HTTP response helpers
├── cmd/
│   ├── migrate/              # Миграции БД
│   └── seed/                 # Seed данные
├── docker-compose.yml        # Docker конфигурация
├── Makefile                  # Команды для сборки и запуска
├── go.mod                    # Go модули
└── README.md                 # Документация
\`\`\`

## Быстрый старт

### Предварительные требования

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 16 (если запускаете локально)
- Redis 7
- Make (опционально)

### Установка

1. Клонируйте репозиторий:
\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
\`\`\`

2. Скопируйте файл окружения:
\`\`\`bash
cp .env.example .env
\`\`\`

3. Отредактируйте `.env` файл и добавьте свои настройки:
\`\`\`env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=quicknode_clone

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT Secret (измените в production!)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Blockchain Node URLs (добавьте свои)
ETH_NODE_URL=https://mainnet.infura.io/v3/YOUR-PROJECT-ID
POLYGON_NODE_URL=https://polygon-rpc.com
\`\`\`

### Запуск с Docker Compose (рекомендуется)

1. Запустите все сервисы:
\`\`\`bash
docker-compose up -d
\`\`\`

2. Проверьте статус:
\`\`\`bash
docker-compose ps
\`\`\`

3. Запустите миграции:
\`\`\`bash
docker-compose exec api-gateway go run cmd/migrate/main.go up
\`\`\`

4. Загрузите тестовые данные:
\`\`\`bash
docker-compose exec api-gateway go run cmd/seed/main.go
\`\`\`

### Запуск локально (для разработки)

1. Запустите инфраструктуру (PostgreSQL, Redis, RabbitMQ):
\`\`\`bash
docker-compose up postgres redis rabbitmq -d
\`\`\`

2. Установите зависимости:
\`\`\`bash
go mod download
\`\`\`

3. Запустите миграции:
\`\`\`bash
go run cmd/migrate/main.go up
\`\`\`

4. Загрузите seed данные:
\`\`\`bash
go run cmd/seed/main.go
\`\`\`

5. Запустите каждый сервис в отдельном терминале:

\`\`\`bash
# Terminal 1 - Auth Service
make run-auth

# Terminal 2 - User Service
make run-user

# Terminal 3 - Blockchain Service
make run-blockchain

# Terminal 4 - Analytics Service
make run-analytics

# Terminal 5 - Billing Service
make run-billing

# Terminal 6 - API Gateway
make run-gateway
\`\`\`

Или используйте Makefile команды для сборки всех сервисов:
\`\`\`bash
make build
\`\`\`

## API Endpoints

### Аутентификация

#### Регистрация
\`\`\`bash
curl -X POST http://localhost:8080/api/v1/auth/register \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
\`\`\`

#### Вход
\`\`\`bash
curl -X POST http://localhost:8080/api/v1/auth/login \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
\`\`\`

Ответ:
\`\`\`json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
\`\`\`

### Защищенные endpoints (требуют Bearer token)

#### Получить профиль
\`\`\`bash
curl -X GET http://localhost:8080/api/v1/user/profile \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
\`\`\`

#### Список blockchain нод
\`\`\`bash
curl -X GET http://localhost:8080/api/v1/blockchain/nodes \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
\`\`\`

#### Статистика использования
\`\`\`bash
curl -X GET http://localhost:8080/api/v1/analytics/usage \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
\`\`\`

#### Создать API ключ
\`\`\`bash
curl -X POST http://localhost:8080/api/v1/api-keys \\
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "My API Key",
    "description": "Production key"
  }'
\`\`\`

## Планы подписок

| План | Запросов/месяц | Цена |
|------|----------------|------|
| Free | 10,000 | $0 |
| Basic | 100,000 | $29.99 |
| Professional | 1,000,000 | $99.99 |
| Enterprise | 10,000,000 | $499.99 |

## Поддерживаемые блокчейны

- Ethereum (Mainnet, Testnets)
- Polygon (Mainnet, Mumbai)
- Binance Smart Chain (BSC)
- Avalanche
- Solana

Вы можете добавить свои blockchain ноды через Blockchain Service.

## Разработка

### Генерация protobuf файлов

Если вы изменили `.proto` файлы:
\`\`\`bash
make proto
\`\`\`

### Запуск тестов
\`\`\`bash
make test
\`\`\`

### Сборка проекта
\`\`\`bash
make build
\`\`\`

### Очистка
\`\`\`bash
make clean
\`\`\`

## Конфигурация blockchain нод

Вы можете добавить свои blockchain ноды (например, ваш собственный TronFullNode) в базу данных через seed скрипт или API:

\`\`\`go
node := &models.BlockchainNode{
    Name:     "My Tron Node",
    Type:     "tron",
    Network:  "mainnet",
    URL:      "http://your-tron-node:8090",
    IsActive: true,
    Priority: 100,
}
\`\`\`

## Production Deployment

### Важные настройки для production:

1. **Измените JWT Secret** в `.env`
2. **Настройте SSL/TLS** для всех сервисов
3. **Включите PostgreSQL SSL** mode
4. **Настройте firewall** правила
5. **Используйте секреты** вместо .env файлов
6. **Настройте мониторинг** (Prometheus, Grafana)
7. **Добавьте логирование** (ELK Stack)
8. **Настройте backup** для PostgreSQL

### Kubernetes Deployment

Для production рекомендуется использовать Kubernetes. Добавьте k8s манифесты в папку `k8s/`.

## Мониторинг и логирование

- Все сервисы пишут логи в stdout
- **Асинхронное логирование** через буферизованные каналы (10k записей)
- Analytics Service собирает метрики каждого запроса
- Можно интегрировать с Prometheus и Grafana

## Горутины и конкурентность

Проект активно использует **горутины** для повышения производительности:

### 🚀 Реализованные компоненты:

1. **Async Logger** (`pkg/async/logger.go`)
   - Асинхронное логирование в БД
   - Буферизованные каналы (10,000 записей)
   - 5 worker горутин с auto-retry

2. **Worker Pool** (`pkg/async/worker_pool.go`)
   - Универсальный пул для выполнения задач
   - Динамическое масштабирование воркеров
   - Статистика и health checks

3. **Parallel Requester** (`pkg/async/parallel_requester.go`)
   - Параллельные запросы к blockchain нодам
   - Автоматический failover
   - Режимы: fastest, failover, all, retry, batch

4. **Async Email Service** (`pkg/email/email.go`)
   - Фоновая отправка email
   - Очередь на 1,000 писем
   - 5 worker горутин

### 📈 Улучшения производительности:

- **До:** Логирование блокировало request на ~100ms
- **После:** Логирование асинхронное, 0ms блокировки
- **До:** Email блокировал response на ~200-500ms
- **После:** Email отправка асинхронная, 0ms блокировки
- **До:** Failover = последовательные запросы
- **После:** Failover = параллельные запросы, минимальная латентность

**Подробнее:** [Документация по горутинам](docs/GOROUTINES.md)

## Troubleshooting

### Проблемы с подключением к БД

Проверьте что PostgreSQL запущен и доступен:
\`\`\`bash
docker-compose logs postgres
\`\`\`

### gRPC ошибки

Убедитесь что все сервисы запущены:
\`\`\`bash
docker-compose ps
\`\`\`

### Rate Limit ошибки

По умолчанию установлен лимит 100 запросов/минуту. Измените в `.env`:
\`\`\`env
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1m
\`\`\`

## Contributing

1. Fork проект
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## Лицензия

MIT License

## Контакты

Для вопросов и предложений создайте Issue в репозитории.

---

**Примечание**: Это demo проект для образовательных целей. Не используйте в production без дополнительной настройки безопасности и тестирования.
