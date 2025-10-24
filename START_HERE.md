# ✅ API Работает БЕЗ ПОРТА!

## 🎉 API запущен на: `http://localhost`

**Без :8080!** Теперь можно обращаться просто:
- ✅ `http://localhost` - Web документация (автоматический редирект на /docs)
- ✅ `http://localhost/docs` - Полная интерактивная документация API на русском
- ✅ `http://localhost/health` - Health check
- ✅ `http://localhost/api/v1/auth/login` - Login
- ✅ `http://localhost/api/v1/blockchain/nodes` - Blockchain ноды

---

## 🚀 Как тестировать в Postman:

### 1. Импорт коллекции

1. Откройте **Postman**
2. Нажмите **Import**
3. Выберите файл: **`IronNode.postman_collection.json`**

### 2. Тест последовательность

#### Шаг 1: Health Check
```
GET http://localhost/health
```
Ответ: `{"status":"ok","message":"API is running"}`

#### Шаг 2: Login (demo user)
```
POST http://localhost/api/v1/auth/login

Body:
{
  "email": "demo@example.com",
  "password": "password123"
}
```
✅ Токен автоматически сохранится!

#### Шаг 3: Get Profile
```
GET http://localhost/api/v1/user/profile
Authorization: Bearer {{token}}
```

#### Шаг 4: Blockchain Nodes
```
GET http://localhost/api/v1/blockchain/nodes
Authorization: Bearer {{token}}
```

```
Authorization: Bearer {{token}}

Body:
{
  "jsonrpc": "2.0",
  "method": "eth_blockNumber",
  "params": [],
  "id": 1
}
```

---

## 🔐 Тест восстановления пароля:

#### Шаг 1: Forgot Password
```
POST http://localhost/api/v1/auth/forgot-password

Body:
{
  "email": "demo@example.com"
}
```
✅ Токен сброса автоматически сохранится! (В консоли API увидите ссылку с токеном)

#### Шаг 2: Verify Reset Token
```
POST http://localhost/api/v1/auth/verify-reset-token

Body:
{
  "token": "{{reset_token}}"
}
```

#### Шаг 3: Reset Password
```
POST http://localhost/api/v1/auth/reset-password

Body:
{
  "token": "{{reset_token}}",
  "new_password": "newpassword123"
}
```

#### Шаг 4: Login с новым паролем
```
POST http://localhost/api/v1/auth/login

Body:
{
  "email": "demo@example.com",
  "password": "newpassword123"
}
```

---

## 📝 Все работающие endpoints:

### Публичные (без токена):
- `GET /health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/forgot-password`
- `POST /api/v1/auth/verify-reset-token`
- `POST /api/v1/auth/reset-password`

### Защищенные (с токеном):
- `GET /api/v1/user/profile`
- `GET /api/v1/blockchain/nodes`
- `GET /api/v1/blockchain/nodes/:id`
- `GET /api/v1/analytics/usage`
- `GET /api/v1/analytics/requests`
- `GET /api/v1/api-keys`
- `POST /api/v1/api-keys`
- `DELETE /api/v1/api-keys/:id`

---

## 🎯 Быстрый тест через cURL:

```bash
# Health
curl http://localhost/health

# Login
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"password123"}'

# Get nodes (замените YOUR_TOKEN)
curl http://localhost/api/v1/blockchain/nodes \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 💡 Demo данные:

**Пользователь:**
- Email: `demo@example.com`
- Password: `password123`

**Blockchain ноды:**
- Ethereum Mainnet
- Polygon Mainnet
- BSC Mainnet

---

## 🐛 Если API не работает:

Перезапустите API (требуются права администратора):

```bash
cd "C:\Users\backend\Desktop\Cloud AI\Go"
go run cmd/standalone-api/main.go
```

Появится окно UAC - подтвердите права администратора.

---

## 📊 Что это даёт:

✅ **Без порта** - обращение как `http://localhost` вместо `http://localhost:8080`
✅ **Стандартный HTTP** - работает на порту 80
✅ **Проще URL** - короче и чище
✅ **Production-like** - как настоящие API

---

## 📁 Файлы проекта:

1. **`IronNode.postman_collection.json`** - Postman коллекция (уже настроена без порта!)
2. **`QUICK_START.md`** - подробная инструкция
3. **`POSTMAN_TESTING.md`** - гайд по тестированию
4. **`cmd/standalone-api/main.go`** - API сервер

---

**Статус:** ✅ API работает на `http://localhost` БЕЗ ПОРТА!

Можете сразу тестировать! 🚀
