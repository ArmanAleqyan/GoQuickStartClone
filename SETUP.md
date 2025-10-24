# Руководство по запуску проекта

## Что уже сделано ✅

1. ✅ **Установлены все Go зависимости**
2. ✅ **Запущены Docker контейнеры:**
   - PostgreSQL (порт 5433)
   - Redis (порт 6379)
   - RabbitMQ (порт 5672, управление 15672)
3. ✅ **Выполнены миграции базы данных**
4. ✅ **Загружены тестовые данные:**
   - Demo пользователь: `demo@example.com` / `password123`
   - 3 blockchain ноды (Ethereum, Polygon, BSC)
   - Free подписка для demo пользователя

## Текущий статус

**Docker контейнеры запущены и работают:**

\`\`\`bash
docker-compose ps
\`\`\`

Вы должны увидеть:
- ✅ quicknode_postgres - HEALTHY
- ✅ quicknode_redis - HEALTHY
- ✅ quicknode_rabbitmq - HEALTHY

**База данных настроена:**
- Все таблицы созданы
- Тестовые данные загружены
- Подключение: localhost:5433

## Следующий шаг - Генерация Protobuf файлов

Для запуска gRPC сервисов нужно сгенерировать protobuf файлы.

### Установка protoc (Protocol Buffers compiler)

**Windows:**

1. Скачайте protoc:
   - Перейдите на https://github.com/protocolbuffers/protobuf/releases
   - Скачайте `protoc-<version>-win64.zip`
   - Распакуйте в `C:\protoc`

2. Добавьте в PATH:
   - Откройте "Переменные среды"
   - Добавьте `C:\protoc\bin` в PATH

3. Установите Go плагины:
\`\`\`bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
\`\`\`

4. Убедитесь что Go bin в PATH:
\`\`\`bash
# Добавьте в PATH если нет
set PATH=%PATH%;%USERPROFILE%\go\bin
\`\`\`

### Генерация protobuf файлов

После установки protoc выполните:

\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"

# Auth Service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/auth-service/proto/auth.proto

# User Service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/user-service/proto/user.proto

# Blockchain Service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/blockchain-service/proto/blockchain.proto


# Analytics Service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/analytics-service/proto/analytics.proto

# Billing Service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/billing-service/proto/billing.proto
\`\`\`

Или используйте Makefile:
\`\`\`bash
make proto
\`\`\`

## Запуск сервисов

После генерации protobuf файлов запустите каждый сервис в отдельном терминале:

### Terminal 1 - Auth Service
\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
go run services/auth-service/cmd/main.go
\`\`\`

### Terminal 2 - User Service
\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
go run services/user-service/cmd/main.go
\`\`\`

### Terminal 3 - Blockchain Service
\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
go run services/blockchain-service/cmd/main.go
\`\`\`

\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
\`\`\`

### Terminal 5 - Analytics Service
\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
go run services/analytics-service/cmd/main.go
\`\`\`

### Terminal 6 - Billing Service
\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
go run services/billing-service/cmd/main.go
\`\`\`

### Terminal 7 - API Gateway (главный)
\`\`\`bash
cd "C:/Users/backend/Desktop/Cloud AI/Go"
go run services/api-gateway/cmd/main.go
\`\`\`

## Тестирование API

После запуска API Gateway на порту 8080, вы можете тестировать API:

### 1. Вход с demo пользователем
\`\`\`bash
curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d "{\"email\":\"demo@example.com\",\"password\":\"password123\"}"
\`\`\`

Скопируйте токен из ответа.

### 2. Получить профиль
\`\`\`bash
curl -X GET http://localhost:8080/api/v1/user/profile -H "Authorization: Bearer YOUR_TOKEN"
\`\`\`

### 3. Список blockchain нод
\`\`\`bash
curl -X GET http://localhost:8080/api/v1/blockchain/nodes -H "Authorization: Bearer YOUR_TOKEN"
\`\`\`

## Альтернатива - Использование Docker Compose

Вместо ручного запуска каждого сервиса, вы можете использовать Docker Compose для запуска всех сервисов:

\`\`\`bash
# Сначала сгенерируйте protobuf файлы (см. выше)

# Затем запустите все сервисы
docker-compose up -d --build
\`\`\`

## Проверка статуса

\`\`\`bash
# Проверить Docker контейнеры
docker-compose ps

# Проверить логи
docker-compose logs -f api-gateway
docker-compose logs -f auth-service

# Остановить все
docker-compose down
\`\`\`

## Полезные команды

\`\`\`bash
# Пересоздать миграции
go run cmd/migrate/main.go down
go run cmd/migrate/main.go up

# Перезагрузить seed данные
go run cmd/seed/main.go

# Собрать все сервисы
make build

# Запустить тесты
make test

# Очистить
make clean
\`\`\`

## Структура портов

| Сервис | Порт | Тип |
|--------|------|-----|
| API Gateway | 8080 | HTTP |
| Auth Service | 50051 | gRPC |
| User Service | 50052 | gRPC |
| Blockchain Service | 50053 | gRPC |
| Analytics Service | 50055 | gRPC |
| Billing Service | 50056 | gRPC |
| PostgreSQL | 5433 | TCP |
| Redis | 6379 | TCP |
| RabbitMQ | 5672 | TCP |
| RabbitMQ Management | 15672 | HTTP |

## Troubleshooting

### Проблема: "protoc not found"
**Решение:** Установите protoc (см. инструкции выше)

### Проблема: "package proto is not in std"
**Решение:** Сгенерируйте protobuf файлы командой `make proto`

### Проблема: "failed to connect to database"
**Решение:** Убедитесь что PostgreSQL запущен на порту 5433:
\`\`\`bash
docker-compose ps postgres
\`\`\`

### Проблема: Порт уже занят
**Решение:** Измените порты в `.env` файле

## Следующие шаги

1. Установите protoc
2. Сгенерируйте protobuf файлы
3. Запустите все сервисы
4. Протестируйте API
5. Добавьте свои blockchain ноды в базу данных
6. Настройте production конфигурацию

Enjoy! 🚀
