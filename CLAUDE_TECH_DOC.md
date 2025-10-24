# 🤖 ТЕХНИЧЕСКАЯ ДОКУМЕНТАЦИЯ ДЛЯ CLAUDE
## Проект: IronNode - Blockchain Node Infrastructure Platform

> **Цель документа**: Быстрая ориентация в проекте для выполнения задач без долгого изучения кода

---

## 📋 КРАТКОЕ ОПИСАНИЕ

**IronNode** - это микросервисная платформа для доступа к blockchain нодам через REST API.

**Технологии**: Go 1.23, Gin, gRPC, PostgreSQL, Redis, RabbitMQ, Docker

**Архитектура**: 6 микросервисов + API Gateway + инфраструктура

---

## 🏗️ СТРУКТУРА ПРОЕКТА

```
C:\Users\backend\Desktop\Cloud AI\Go\
├── services/              # Все микросервисы
│   ├── api-gateway/      # HTTP API Gateway (порт 8080)
│   ├── auth-service/     # Аутентификация (gRPC :50051)
│   ├── user-service/     # Управление пользователями (gRPC :50052)
│   ├── blockchain-service/  # Blockchain ноды (gRPC :50053)
│   ├── analytics-service/   # Аналитика (gRPC :50055)
│   └── billing-service/     # Биллинг (gRPC :50056)
│
├── pkg/                  # Общие пакеты
│   ├── config/          # Конфигурация (.env загрузка)
│   ├── database/        # PostgreSQL подключение
│   ├── cache/           # Redis клиент
│   ├── models/          # Database модели (User, APIKey, etc)
│   ├── middleware/      # HTTP middleware (CORS, RateLimit, Auth)
│   ├── logger/          # Логирование
│   ├── async/           # Асинхронные компоненты
│   ├── email/           # Email сервис
│   └── response/        # HTTP response helpers
│
├── cmd/                 # CLI команды
│   ├── migrate/         # Миграции БД
│   ├── seed/            # Seed данные
│   └── standalone-api/  # Standalone версия API
│
├── .env                 # Конфигурация окружения
├── docker-compose.yml   # Docker конфигурация
├── Makefile            # Команды сборки
└── go.mod              # Go зависимости
```

---

## 🎯 МИКРОСЕРВИСЫ - ДЕТАЛЬНОЕ ОПИСАНИЕ

### 1️⃣ API GATEWAY (порт 80 → 8080)
**Файл**: `services/api-gateway/cmd/main.go`
**Роль**: Единая точка входа для всех HTTP запросов

**Основные компоненты**:
- `internal/routes/routes.go` - Все маршруты API
- `internal/handler/auth_handler.go` - Обработка auth запросов
- `internal/handler/blockchain_handler.go` - Обработка blockchain запросов
- `internal/handler/analytics_handler.go` - Обработка analytics запросов
- `internal/handler/api_key_handler.go` - Обработка API ключей

**Ключевые зависимости**:
- Gin (HTTP framework)
- Redis (rate limiting)
- gRPC клиенты для всех сервисов

**Middleware**:
- CORS (`pkg/middleware/cors.go`)
- Rate Limiter (`pkg/middleware/ratelimit.go`)
- Auth Middleware (JWT валидация)
- Request Logger (`pkg/middleware/request_logger.go`)

---

### 2️⃣ AUTH SERVICE (gRPC :50051)
**Файл**: `services/auth-service/cmd/main.go`
**Роль**: Аутентификация и авторизация пользователей

**Основные функции**:
- Регистрация пользователей (`Register`)
- Логин (`Login`)
- Генерация JWT токенов
- Валидация токенов (`ValidateToken`)
- Сброс пароля (`ForgotPassword`, `ResetPassword`)

**Service Layer**: `services/auth-service/internal/service/auth_service.go`
- `Register(email, password, firstName, lastName)` - создание пользователя
- `Login(email, password)` - возвращает JWT токен
- `ValidateToken(tokenString)` - проверка токена
- `ForgotPassword(email)` - генерирует reset token
- `ResetPassword(token, newPassword)` - сброс пароля

**Repository**: `services/auth-service/internal/repository/auth_repository.go`
- CRUD операции для User и PasswordReset

**Технологии**:
- bcrypt (хеширование паролей)
- JWT (`github.com/golang-jwt/jwt/v5`)
- Email сервис (асинхронная отправка)

---

### 3️⃣ USER SERVICE (gRPC :50052)
**Файл**: `services/user-service/cmd/main.go`
**Роль**: Управление API ключами пользователей

**Основные функции**:
- Создание API ключей
- Список API ключей пользователя
- Валидация API ключей
- Удаление API ключей

**Service Layer**: `services/user-service/internal/service/user_service.go`
**Repository**: `services/user-service/internal/repository/user_repository.go`

**Модель**: `pkg/models/api_key.go`
- Связь с User через UserID
- Проверка на истечение срока действия
- Флаг IsActive

---

### 4️⃣ BLOCKCHAIN SERVICE (gRPC :50053)
**Файл**: `services/blockchain-service/cmd/main.go`
**Роль**: Управление blockchain нодами

**Основные функции**:
- Список доступных нод
- Получение ноды по ID
- Создание/обновление нод
- Приоритизация нод

**Service Layer**: `services/blockchain-service/internal/service/node_service.go`
**Repository**: `services/blockchain-service/internal/repository/node_repository.go`

**Модель**: `pkg/models/blockchain_node.go`
- Поддерживаемые типы: ethereum, bitcoin, polygon, bsc, avalanche, solana
- Priority - для выбора предпочтительной ноды
- MaxRequests - лимит запросов

---

### 5️⃣ ANALYTICS SERVICE (gRPC :50055)
**Файл**: `services/analytics-service/cmd/main.go`
**Роль**: Логирование запросов и аналитика

**Основные функции**:
- Логирование каждого запроса (асинхронно)
- Получение статистики использования
- История запросов пользователя
- Метрики производительности

**Service Layer**: `services/analytics-service/internal/service/analytics_service.go`
**Repository**: `services/analytics-service/internal/repository/analytics_repository.go`

**Модель**: `pkg/models/request_log.go`
- UserID, APIKeyID
- Blockchain, Method, Endpoint
- StatusCode, ResponseTime
- RequestSize, ResponseSize

**Асинхронное логирование**: `pkg/async/logger.go`
- Буферизованный канал на 10,000 записей
- 5 worker горутин
- Retry логика (3 попытки)

---

### 6️⃣ BILLING SERVICE (gRPC :50056)
**Файл**: `services/billing-service/cmd/main.go`
**Роль**: Управление подписками и квотами

**Основные функции**:
- Управление подписками
- Проверка квот
- Отслеживание использования

**Service Layer**: `services/billing-service/internal/service/billing_service.go`
**Repository**: `services/billing-service/internal/repository/billing_repository.go`

**Модель**: `pkg/models/subscription.go`
- PlanType: free, basic, professional, enterprise
- RequestsPerMonth, RequestsUsed
- Методы: `HasRequestsAvailable()`, `IsExpired()`

**Планы подписок**:
- Free: 10,000 req/month, $0
- Basic: 100,000 req/month, $29.99
- Professional: 1,000,000 req/month, $99.99
- Enterprise: 10,000,000 req/month, $499.99

---

## 📦 ОБЩИЕ ПАКЕТЫ (pkg/)

### Config (`pkg/config/config.go`)
**Назначение**: Загрузка конфигурации из .env

**Структуры**:
```go
type Config struct {
    Environment string
    Database    DatabaseConfig
    Redis       RedisConfig
    RabbitMQ    RabbitMQConfig
    JWT         JWTConfig
    Services    ServicesConfig
    Email       EmailConfig
}
```

**Использование**:
```go
cfg, err := config.Load()
dsn := cfg.Database.DSN()
```

---

### Database (`pkg/database/postgres.go`)
**Назначение**: Подключение к PostgreSQL через GORM

**Функция**: `NewPostgresConnection(dsn string) (*gorm.DB, error)`

---

### Models (`pkg/models/`)
**Все модели базы данных**:

1. **User** (`user.go`)
   - ID (uuid), Email, Password (hashed), FirstName, LastName
   - IsActive, CreatedAt, UpdatedAt

2. **APIKey** (`api_key.go`)
   - ID, UserID, Key, Name, Description
   - IsActive, ExpiresAt
   - Метод: `IsExpired() bool`

3. **BlockchainNode** (`blockchain_node.go`)
   - ID, Name, Type, Network, URL
   - IsActive, Priority, MaxRequests

4. **Subscription** (`subscription.go`)
   - ID, UserID, PlanType
   - RequestsPerMonth, RequestsUsed, Price
   - StartsAt, EndsAt
   - Методы: `HasRequestsAvailable()`, `IsExpired()`

5. **RequestLog** (`request_log.go`)
   - ID, UserID, APIKeyID, Blockchain
   - Method, Endpoint, StatusCode
   - ResponseTime, RequestSize, ResponseSize

6. **PasswordReset** (`password_reset.go`)
   - ID, UserID, Token
   - ExpiresAt, UsedAt
   - Метод: `IsValid() bool`

---

### Middleware (`pkg/middleware/`)

1. **CORS** (`cors.go`)
   - Разрешает cross-origin запросы

2. **RateLimiter** (`ratelimit.go`)
   - Redis-based rate limiting
   - По умолчанию: 100 запросов/минуту
   - Использование: `rateLimiter.Limit()`

3. **RequestLogger** (`request_logger.go`)
   - Логирование HTTP запросов

---

### Async (`pkg/async/`)

1. **AsyncLogger** (`logger.go`)
   - Асинхронное логирование в БД
   - 10,000 буфер + 5 workers
   - Retry логика
   - Методы: `Log()`, `Shutdown()`, `Stats()`

2. **WorkerPool** (`worker_pool.go`)
   - Универсальный пул воркеров
   - Динамическое масштабирование

3. **ParallelRequester** (`parallel_requester.go`)
   - Параллельные HTTP запросы
   - Failover между нодами

---

### Email (`pkg/email/email.go`)
**Назначение**: Асинхронная отправка email

**Компоненты**:
- Буферизованная очередь (1000 писем)
- 5 worker горутин
- Graceful shutdown

**Методы**:
- `SendPasswordResetEmail(email, token, resetURL)`
- `SendWelcomeEmail(email, firstName)`
- `SendPasswordChangedEmail(email)`

---

## 🔌 API ENDPOINTS

### Публичные (без авторизации)

#### POST /api/v1/auth/register
```bash
curl -X POST http://localhost/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### POST /api/v1/auth/login
```bash
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```
**Ответ**: `{"token": "eyJhbGc..."}`

#### POST /api/v1/auth/forgot-password
```bash
curl -X POST http://localhost/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

#### POST /api/v1/auth/verify-reset-token
#### POST /api/v1/auth/reset-password

---

### Защищенные (требуют JWT Bearer token)

**Заголовок**: `Authorization: Bearer YOUR_JWT_TOKEN`

#### GET /api/v1/user/profile
Получить профиль текущего пользователя

#### GET /api/v1/blockchain/nodes
Список всех blockchain нод

#### GET /api/v1/blockchain/nodes/:id
Получить ноду по ID

#### GET /api/v1/analytics/usage
Статистика использования

#### GET /api/v1/analytics/requests
История запросов

#### GET /api/v1/api-keys
Список API ключей

#### POST /api/v1/api-keys
Создать новый API ключ

#### DELETE /api/v1/api-keys/:id
Удалить API ключ

---

## 🗄️ БАЗА ДАННЫХ

**PostgreSQL 16** (порт 5433 снаружи, 5432 внутри контейнера)

### Миграции

**Запуск миграций**:
```bash
go run cmd/migrate/main.go up
```

**Откат миграций**:
```bash
go run cmd/migrate/main.go down
```

**Файл**: `cmd/migrate/main.go`

**Мигрируемые таблицы**:
- users
- api_keys
- blockchain_nodes
- request_logs
- subscriptions
- password_resets

---

### Seed данные

**Запуск**:
```bash
go run cmd/seed/main.go
```

**Файл**: `cmd/seed/main.go`

---

## 🐳 DOCKER

### Инфраструктура

**Файл**: `docker-compose.yml`

**Сервисы**:
1. **postgres** - PostgreSQL 16 (порт 5433)
2. **redis** - Redis 7 (порт 6379)
3. **rabbitmq** - RabbitMQ с management (порты 5672, 15672)
4. **api-gateway** - HTTP API (порт 80)
5. **auth-service** - gRPC (порт 50051)
6. **user-service** - gRPC (порт 50052)
7. **blockchain-service** - gRPC (порт 50053)
8. **analytics-service** - gRPC (порт 50055)
9. **billing-service** - gRPC (порт 50056)

### Команды Docker

```bash
# Запуск всех сервисов
docker-compose up -d

# Пересборка и запуск
docker-compose up -d --build

# Остановка
docker-compose down

# Остановка с удалением volumes
docker-compose down -v

# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f api-gateway

# Статус контейнеров
docker-compose ps
```

---

## 🔧 КОНФИГУРАЦИЯ (.env)

**Файл**: `.env` (создан на основе `.env.example`)

```env
# Database
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ironnode

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRY=24h

# Service Ports
API_GATEWAY_PORT=80
AUTH_SERVICE_PORT=50051
USER_SERVICE_PORT=50052
BLOCKCHAIN_SERVICE_PORT=50053
ANALYTICS_SERVICE_PORT=50055
BILLING_SERVICE_PORT=50056

# Blockchain Nodes
ETH_NODE_URL=https://mainnet.infura.io/v3/YOUR-PROJECT-ID

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Environment
ENVIRONMENT=development
```

---

## 🚀 АСИНХРОННЫЕ КОМПОНЕНТЫ

### AsyncLogger
**Файл**: `pkg/async/logger.go`

**Назначение**: Асинхронное логирование запросов в БД без блокировки request'ов

**Характеристики**:
- Буфер: 10,000 записей
- Воркеры: 5 горутин
- Retry: 3 попытки с exponential backoff

**Использование**:
```go
logger := async.NewAsyncLogger(db, 10000, 5)
logger.Log(requestLog)
logger.Shutdown(5 * time.Second)
```

---

### EmailService
**Файл**: `pkg/email/email.go`

**Назначение**: Асинхронная отправка email

**Характеристики**:
- Буфер: 1,000 писем
- Воркеры: 5 горутин

**Использование**:
```go
emailService := email.NewEmailService("noreply@ironnode.com")
emailService.SendPasswordResetEmail(email, token, resetURL)
emailService.Shutdown()
```

---

## 📝 ТИПИЧНЫЕ ЗАДАЧИ И ГДЕ ИСКАТЬ КОД

### ❓ "Добавить новый endpoint в API"
1. **Routes**: `services/api-gateway/internal/routes/routes.go` - добавить маршрут
2. **Handler**: Создать handler в `services/api-gateway/internal/handler/`
3. **Service**: Добавить gRPC вызов к нужному сервису

---

### ❓ "Изменить логику аутентификации"
**Файлы**:
- Service: `services/auth-service/internal/service/auth_service.go`
- Handler: `services/auth-service/internal/handler/auth_handler.go`
- Repository: `services/auth-service/internal/repository/auth_repository.go`

---

### ❓ "Добавить новую модель в БД"
1. Создать модель в `pkg/models/`
2. Добавить в `cmd/migrate/main.go` в `runMigrations()`
3. Запустить `go run cmd/migrate/main.go up`

---

### ❓ "Изменить rate limit"
**Файл**: `services/api-gateway/internal/routes/routes.go`
```go
rateLimiter := middleware.NewRateLimiter(redisClient, 100, 1*time.Minute)
//                                                    ^^^  ^^^^^^^^^^^^
//                                                 requests   window
```

---

### ❓ "Добавить новый тип blockchain"
**Файл**: `pkg/models/blockchain_node.go`
```go
const (
    Ethereum BlockchainType = "ethereum"
    Bitcoin  BlockchainType = "bitcoin"
    // Добавить здесь
)
```

---

### ❓ "Изменить JWT expiry"
**Файл**: `.env`
```env
JWT_EXPIRY=24h  # Изменить на нужное значение
```

---

### ❓ "Добавить middleware"
1. Создать в `pkg/middleware/`
2. Применить в `services/api-gateway/internal/routes/routes.go`

---

### ❓ "Настроить SMTP для email"
**Файл**: `pkg/email/email.go`
- Изменить функцию `sendEmail()` для использования SMTP/SendGrid

---

## 🧪 ТЕСТИРОВАНИЕ

### Postman Collection
**Файл**: `IronNode.postman_collection.json`
- Импортировать в Postman для тестирования API

### Документация
- **QUICK_START.md** - Быстрый старт
- **SETUP.md** - Подробная настройка
- **POSTMAN_TESTING.md** - Тестирование в Postman
- **SENIOR_REVIEW.md** - Code review

---

## 🔍 ЗАВИСИМОСТИ (go.mod)

```go
require (
    github.com/gin-gonic/gin v1.9.1          // HTTP framework
    github.com/golang-jwt/jwt/v5 v5.2.0      // JWT
    github.com/google/uuid v1.5.0            // UUID
    github.com/redis/go-redis/v9 v9.4.0      // Redis client
    golang.org/x/crypto v0.18.0              // bcrypt
    google.golang.org/grpc v1.60.1           // gRPC
    gorm.io/driver/postgres v1.5.4           // PostgreSQL driver
    gorm.io/gorm v1.25.5                     // ORM
)
```

---

## 🔐 БЕЗОПАСНОСТЬ

### JWT Authentication
- Secret: настраивается в `.env`
- Expiry: 24 часа по умолчанию
- Claims: UserID, Email

### Password Hashing
- bcrypt с DefaultCost (10)

### Rate Limiting
- Redis-based
- 100 req/min по умолчанию
- Per user или per IP

---

## 🎯 БЫСТРЫЕ КОМАНДЫ

### Запуск проекта
```bash
docker-compose up -d --build
```

### Миграции
```bash
go run cmd/migrate/main.go up
```

### Seed
```bash
go run cmd/seed/main.go
```

### Логи
```bash
docker-compose logs -f api-gateway
```

### Остановка
```bash
docker-compose down
```

---

## 📊 АРХИТЕКТУРНАЯ ДИАГРАММА

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────────────────────────────────┐
│         API Gateway (:80)               │
│  - Routes                               │
│  - Middleware (CORS, Rate Limit, Auth)  │
└────┬────┬────┬────┬────┬────────────────┘
     │    │    │    │    │
     │gRPC│gRPC│gRPC│gRPC│gRPC
     ▼    ▼    ▼    ▼    ▼
┌────────┐ ┌────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐
│  Auth  │ │  User  │ │Blockchain│ │Analytics │ │ Billing │
│Service │ │Service │ │ Service  │ │ Service  │ │ Service │
│:50051  │ │:50052  │ │  :50053  │ │  :50055  │ │ :50056  │
└───┬────┘ └───┬────┘ └────┬─────┘ └────┬─────┘ └────┬────┘
    │          │           │            │            │
    └──────────┴───────────┴────────────┴────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │   PostgreSQL :5433   │
                └──────────────────────┘
                │   Redis :6379        │
                └──────────────────────┘
                │   RabbitMQ :5672     │
                └──────────────────────┘
```

---

## ✅ ЧЕКЛИСТ ДЛЯ ИЗМЕНЕНИЙ

При добавлении нового функционала:

1. ☐ Определить затронутые сервисы
2. ☐ Обновить proto файлы (если gRPC)
3. ☐ Добавить/изменить модели в `pkg/models/`
4. ☐ Обновить миграции в `cmd/migrate/main.go`
5. ☐ Создать/обновить repository
6. ☐ Создать/обновить service
7. ☐ Создать/обновить handler
8. ☐ Добавить routes в `api-gateway`
9. ☐ Протестировать через Postman
10. ☐ Обновить документацию

---

## 🎓 КЛЮЧЕВЫЕ КОНЦЕПЦИИ

### Clean Architecture
```
Handler → Service → Repository → Database
```

### gRPC Communication
- Все сервисы общаются через gRPC
- API Gateway - единственный HTTP endpoint

### Async Processing
- Email отправка - асинхронная
- Логирование запросов - асинхронное

### Rate Limiting
- На уровне API Gateway
- Использует Redis

---

## 📞 TROUBLESHOOTING

### Проблема: Не могу подключиться к БД
**Решение**: Проверить контейнер PostgreSQL
```bash
docker-compose logs postgres
docker-compose ps
```

### Проблема: gRPC connection refused
**Решение**: Убедиться что все сервисы запущены
```bash
docker-compose ps
docker-compose up -d
```

### Проблема: Rate limit exceeded
**Решение**: Изменить лимит в routes.go или подождать 1 минуту

---

## 🎯 ЗАКЛЮЧЕНИЕ

Эта документация содержит всё необходимое для:
- ✅ Быстрого понимания архитектуры
- ✅ Навигации по коду
- ✅ Выполнения типичных задач
- ✅ Отладки проблем
- ✅ Добавления нового функционала

**При получении задачи**:
1. Определить затронутые компоненты по этой документации
2. Найти нужные файлы
3. Внести изменения
4. Протестировать

---

**Последнее обновление**: 2025-10-17
**Версия проекта**: IronNode v1.0
**Go версия**: 1.23
