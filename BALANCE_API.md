# 💰 Balance API - Получение балансов TRX и USDT

> API для получения балансов Tron (TRX) и USDT TRC20 токенов

---

## 📋 Описание

Balance API позволяет получать актуальные балансы кошельков в сети Tron:
- **TRX** - нативная криптовалюта Tron
- **USDT TRC20** - стейблкоин USDT на сети Tron

Все запросы идут напрямую к Tron Full Node: `http://78.46.94.60:8090`

---

## 🔐 Авторизация

Все endpoints требуют JWT токен в заголовке:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

---

## 📡 API Endpoints

### 1. Получить оба баланса (TRX + USDT)

**GET** `/api/v1/balance/tron/:address`

Получает балансы TRX и USDT для указанного адреса.

#### Параметры

| Параметр | Тип | Описание |
|----------|-----|----------|
| address | string (path) | Tron адрес (начинается с 'T', 34 символа) |

#### Пример запроса

```bash
GET http://localhost/api/v1/balance/tron/TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Пример ответа

```json
{
  "success": true,
  "message": "Balances retrieved successfully",
  "data": {
    "address": "TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz",
    "trx_balance": "1500000",
    "usdt_balance": "100000000",
    "trx_decimal": "1.500000",
    "usdt_decimal": "100.000000"
  }
}
```

#### Поля ответа

| Поле | Описание |
|------|----------|
| address | Tron адрес кошелька |
| trx_balance | Баланс TRX в SUN (1 TRX = 1,000,000 SUN) |
| usdt_balance | Баланс USDT в минимальных единицах (1 USDT = 1,000,000) |
| trx_decimal | Баланс TRX в человекочитаемом формате |
| usdt_decimal | Баланс USDT в человекочитаемом формате |

---

### 2. Получить только баланс TRX

**GET** `/api/v1/balance/trx/:address`

Получает только баланс TRX для указанного адреса.

#### Пример запроса

```bash
GET http://localhost/api/v1/balance/trx/TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Пример ответа

```json
{
  "success": true,
  "message": "TRX balance retrieved successfully",
  "data": {
    "address": "TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz",
    "trx_balance": "1500000",
    "trx_decimal": "1500000"
  }
}
```

---

### 3. Получить только баланс USDT

**GET** `/api/v1/balance/usdt/:address`

Получает только баланс USDT TRC20 для указанного адреса.

#### Пример запроса

```bash
GET http://localhost/api/v1/balance/usdt/TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Пример ответа

```json
{
  "success": true,
  "message": "USDT balance retrieved successfully",
  "data": {
    "address": "TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz",
    "usdt_balance": "100000000",
    "usdt_decimal": "100000000"
  }
}
```

---

### 4. Получить балансы (POST версия)

**POST** `/api/v1/balance/check`

Получает балансы TRX и USDT через POST запрос.

#### Body параметры

```json
{
  "address": "TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz"
}
```

#### Пример запроса

```bash
POST http://localhost/api/v1/balance/check
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json

{
  "address": "TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz"
}
```

#### Пример ответа

Аналогичен ответу для `/balance/tron/:address`

---

## 💡 Примеры использования

### cURL

```bash
# Получить оба баланса
curl -X GET "http://localhost/api/v1/balance/tron/TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Получить только TRX
curl -X GET "http://localhost/api/v1/balance/trx/TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Получить только USDT
curl -X GET "http://localhost/api/v1/balance/usdt/TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# POST версия
curl -X POST "http://localhost/api/v1/balance/check" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address": "TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz"}'
```

### JavaScript (Fetch)

```javascript
const getBalance = async (address) => {
  const response = await fetch(
    `http://localhost/api/v1/balance/tron/${address}`,
    {
      headers: {
        'Authorization': `Bearer ${yourJWTToken}`
      }
    }
  );

  const data = await response.json();

  if (data.success) {
    console.log('TRX Balance:', data.data.trx_decimal, 'TRX');
    console.log('USDT Balance:', data.data.usdt_decimal, 'USDT');
  }
};

// Использование
getBalance('TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz');
```

### Python (requests)

```python
import requests

def get_balance(address, token):
    url = f"http://localhost/api/v1/balance/tron/{address}"
    headers = {
        "Authorization": f"Bearer {token}"
    }

    response = requests.get(url, headers=headers)
    data = response.json()

    if data["success"]:
        print(f"TRX Balance: {data['data']['trx_decimal']} TRX")
        print(f"USDT Balance: {data['data']['usdt_decimal']} USDT")

    return data

# Использование
balance = get_balance("TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz", "your_jwt_token")
```

---

## ⚠️ Важные замечания

### Валидация адресов

Все адреса проверяются на валидность:
- Должны начинаться с буквы **'T'**
- Длина должна быть **34 символа**
- Формат: Base58

### Конвертация единиц

**TRX:**
- 1 TRX = 1,000,000 SUN
- API возвращает баланс в обоих форматах

**USDT:**
- 1 USDT = 1,000,000 минимальных единиц
- API возвращает баланс в обоих форматах

### Производительность

- Запросы идут напрямую к Tron Full Node
- Среднее время ответа: **100-500ms**
- Рекомендуется кешировать результаты на стороне клиента

### Ошибки

#### 400 Bad Request - Невалидный адрес
```json
{
  "success": false,
  "message": "Invalid Tron address. Address must start with 'T' and be 34 characters long",
  "data": null
}
```

#### 401 Unauthorized - Нет токена
```json
{
  "success": false,
  "message": "User not authenticated",
  "data": null
}
```

#### 500 Internal Server Error - Ошибка при получении баланса
```json
{
  "success": false,
  "message": "Failed to get balances",
  "error": "connection timeout"
}
```

---

## 🔧 Технические детали

### Архитектура

```
Client Request
     ↓
API Gateway (:80)
     ↓
Balance Handler
     ↓
Tron Client (pkg/tron/client.go)
     ↓
Tron Full Node (http://78.46.94.60:8090)
```

### Используемые методы Tron API

1. **Для TRX баланса:**
   - Endpoint: `/wallet/getaccount`
   - Метод: POST
   - Возвращает: account info с балансом

2. **Для USDT баланса:**
   - Endpoint: `/wallet/triggersmartcontract`
   - Метод: POST
   - Contract: `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t` (USDT TRC20)
   - Function: `balanceOf(address)`

### Используемые библиотеки

- `github.com/fbsobreira/gotron-sdk` - для работы с Tron адресами
- `math/big` - для работы с большими числами
- `encoding/hex` - для hex конвертации

---

## 🚀 Быстрый старт

### 1. Получить JWT токен

```bash
# Login
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "password123"
  }'
```

Ответ содержит `token` - используйте его для запросов к Balance API.

### 2. Получить баланс

```bash
# Замените YOUR_TOKEN на полученный токен
curl -X GET "http://localhost/api/v1/balance/tron/TYsNc4W8K8dLY6j8dVJZ9BFpqPrQvY5xVz" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 📊 Use Cases

### 1. Проверка баланса перед транзакцией

```javascript
const checkBalance = async (address, requiredUSDT) => {
  const response = await fetch(`/api/v1/balance/usdt/${address}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });

  const data = await response.json();
  const balance = parseFloat(data.data.usdt_decimal);

  if (balance >= requiredUSDT) {
    console.log('✅ Достаточно USDT для транзакции');
    return true;
  } else {
    console.log('❌ Недостаточно USDT');
    return false;
  }
};
```

### 2. Мониторинг балансов множества адресов

```python
def monitor_balances(addresses, token):
    balances = {}

    for address in addresses:
        response = requests.get(
            f"http://localhost/api/v1/balance/tron/{address}",
            headers={"Authorization": f"Bearer {token}"}
        )

        data = response.json()
        if data["success"]:
            balances[address] = {
                "trx": data["data"]["trx_decimal"],
                "usdt": data["data"]["usdt_decimal"]
            }

    return balances
```

### 3. Алерты при низком балансе

```javascript
const checkLowBalance = async (address, minTRX, minUSDT) => {
  const data = await getBalance(address);

  const trx = parseFloat(data.data.trx_decimal);
  const usdt = parseFloat(data.data.usdt_decimal);

  if (trx < minTRX) {
    alert(`⚠️ Низкий баланс TRX: ${trx} TRX`);
  }

  if (usdt < minUSDT) {
    alert(`⚠️ Низкий баланс USDT: ${usdt} USDT`);
  }
};
```

---

## 📝 FAQ

**Q: Можно ли получить балансы других токенов TRC20?**
A: В текущей версии поддерживается только USDT. Для других токенов нужно добавить их contract address.

**Q: Как часто обновляются балансы?**
A: Балансы получаются в реальном времени с Tron Full Node.

**Q: Есть ли rate limiting?**
A: Да, применяется общий rate limit API Gateway (100 запросов/минуту по умолчанию).

**Q: Можно ли использовать hex адреса вместо base58?**
A: Нет, API принимает только base58 адреса (начинающиеся с 'T').

---

**Дата создания:** 2025-01-28
**Версия API:** 1.0
**Tron Node:** http://78.46.94.60:8090
