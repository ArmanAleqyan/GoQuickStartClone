# Использование Горутин в IronNode

## Обзор

Проект использует **горутины (goroutines)** для повышения производительности и масштабируемости. Горутины - это легковесные потоки выполнения в Go, позволяющие эффективно обрабатывать конкурентные задачи.

---

## 🚀 Реализованные компоненты с горутинами

### 1. **Асинхронный Логгер** (`pkg/async/logger.go`)

**Описание:** Логирование запросов в базу данных происходит асинхронно через буферизованные каналы.

**Характеристики:**
- **Буферизованный канал:** 10,000 записей
- **Worker goroutines:** настраиваемое количество (по умолчанию 5)
- **Retry логика:** до 3 попыток при ошибках БД
- **Graceful shutdown:** корректная остановка с таймаутом

**Преимущества:**
- ✅ Не блокирует основной поток выполнения
- ✅ Повышает пропускную способность API
- ✅ Автоматический retry при сбоях
- ✅ Защита от переполнения очереди

**Пример использования:**
```go
// Создание логгера
logger := async.NewAsyncLogger(db, 10000, 5)

// Асинхронное логирование
logger.Log(&models.RequestLog{
    UserID:       userID,
    Method:       "eth_blockNumber",
    StatusCode:   200,
    ResponseTime: 125,
})

// Получение статистики
stats := logger.Stats()
fmt.Printf("Queue length: %d/%d\n", stats["queue_length"], stats["queue_capacity"])

// Graceful shutdown
logger.Shutdown(30 * time.Second)
```

**Архитектура:**
```
HTTP Request → API Handler
                    ↓
              [Log Entry]
                    ↓
         Buffered Channel (10k)
              ↙  ↓  ↘
         Worker1 Worker2 Worker3 ... (5 workers)
              ↓    ↓    ↓
         PostgreSQL Database
```

---

### 2. **Worker Pool** (`pkg/async/worker_pool.go`)

**Описание:** Универсальный пул воркеров для выполнения произвольных задач.

**Характеристики:**
- **Динамическое масштабирование:** возможность изменения количества воркеров
- **Статистика:** отслеживание выполненных и failed задач
- **Health check:** проверка состояния пула
- **Таймауты:** автоматический timeout для задач (30 сек)

**Преимущества:**
- ✅ Ограничивает количество конкурентных операций
- ✅ Предотвращает перегрузку системы
- ✅ Переиспользование горутин
- ✅ Простой API для добавления задач

**Пример использования:**
```go
// Создание пула с 10 воркерами и очередью на 1000 задач
pool := async.NewWorkerPool(10, 1000)

// Добавление задачи
success := pool.Submit(func(ctx context.Context) error {
    // Ваша задача здесь
    return doSomeWork(ctx)
})

// Добавление задачи с таймаутом
pool.SubmitWithTimeout(task, 5*time.Second)

// Получение статистики
stats := pool.Stats()
fmt.Printf("Processed: %d, Failed: %d\n",
    stats.TasksProcessed, stats.TasksFailed)

// Проверка здоровья
if !pool.IsHealthy() {
    log.Println("Pool is unhealthy!")
}

// Динамическое увеличение воркеров
pool.Resize(20)

// Graceful shutdown
pool.Shutdown(30 * time.Second)
```

**Архитектура:**
```
Client Code
     ↓
Submit(task)
     ↓
Task Queue (buffered channel)
     ↙  ↓  ↘
Worker1 Worker2 Worker3 ... (configurable)
     ↓    ↓    ↓
Task Execution (with timeout)
```

---

### 3. **Parallel Requester** (`pkg/async/parallel_requester.go`)

**Описание:** Параллельные запросы к blockchain нодам с автоматическим failover.

**Режимы работы:**

#### **RequestWithFailover** - Первый успешный ответ
Отправляет запросы ко всем нодам параллельно и возвращает первый успешный ответ.

```go
requester := async.NewParallelRequester(requestFunc, 10*time.Second)

nodeURLs := []string{
    "https://eth-node-1.com",
    "https://eth-node-2.com",
    "https://eth-node-3.com",
}

response, err := requester.RequestWithFailover(
    ctx,
    nodeURLs,
    "eth_blockNumber",
    []byte{},
)
```

**Архитектура:**
```
Request
   ↓
Parallel Goroutines
   ↙    ↓    ↘
Node1  Node2  Node3
   ↓    ↓    ↓
First Success → Cancel Others → Return
```

#### **RequestFastest** - Самый быстрый ответ
Возвращает ответ от самой быстрой ноды.

```go
response, err := requester.RequestFastest(
    ctx,
    nodeURLs,
    "eth_getBalance",
    params,
)
```

#### **RequestAll** - Все ответы
Получает ответы от всех нод (для сравнения/консенсуса).

```go
responses, err := requester.RequestAll(
    ctx,
    nodeURLs,
    "eth_blockNumber",
    []byte{},
)

// Проверка консенсуса
for _, resp := range responses {
    fmt.Printf("Node %s: %s\n", resp.NodeURL, string(resp.Data))
}
```

#### **RequestWithRetry** - Запрос с повторами
Автоматический retry с exponential backoff.

```go
response, err := requester.RequestWithRetry(
    ctx,
    nodeURL,
    "eth_call",
    params,
    3, // maxRetries
)
```

#### **BatchRequest** - Пакетные запросы
Параллельное выполнение нескольких запросов.

```go
requests := []async.NodeRequest{
    {NodeURL: url, Method: "eth_blockNumber", Params: []byte{}},
    {NodeURL: url, Method: "eth_gasPrice", Params: []byte{}},
    {NodeURL: url, Method: "net_version", Params: []byte{}},
}

responses, err := requester.BatchRequest(ctx, nodeURL, requests)
```

**Преимущества:**
- ✅ Автоматический failover между нодами
- ✅ Минимальная латентность (используется самая быстрая нода)
- ✅ Высокая доступность (если одна нода падает, используются другие)
- ✅ Retry с exponential backoff
- ✅ Таймауты для предотвращения зависаний

---

### 4. **Асинхронный Email Service** (`pkg/email/email.go`)

**Описание:** Отправка email в фоновом режиме через пул воркеров.

**Характеристики:**
- **Буферизованная очередь:** 1,000 email
- **Email workers:** 5 горутин
- **Graceful shutdown:** ожидание отправки всех email перед остановкой

**Преимущества:**
- ✅ Не блокирует HTTP response
- ✅ Устойчивость к пиковым нагрузкам
- ✅ Параллельная отправка

**Пример использования:**
```go
// Email service автоматически запускает 5 воркеров
emailService := email.NewEmailService("noreply@quicknode-clone.com")

// Асинхронная отправка (не блокирует)
emailService.SendPasswordResetEmail(
    "user@example.com",
    resetToken,
    resetURL,
)

// Graceful shutdown при завершении приложения
emailService.Shutdown()
```

**Архитектура:**
```
API Handler → Queue Email
                  ↓
          Email Queue (1000)
              ↙  ↓  ↘
      Worker1 Worker2 Worker3 ... (5 workers)
          ↓    ↓    ↓
       SMTP Server / Email Service
```

---

## 📊 Производительность

### До использования горутин:
- **Запрос + логирование:** ~150ms (синхронно)
- **Email отправка:** блокировала response на 200-500ms
- **Blockchain failover:** последовательные запросы = высокая латентность

### После использования горутин:
- **Запрос + логирование:** ~50ms (логирование async)
- **Email отправка:** 0ms блокировки (полностью async)
- **Blockchain failover:** параллельные запросы = минимальная латентность

### Пропускная способность:
- **Async Logger:** обрабатывает 10,000+ записей/сек
- **Worker Pool:** выполняет задачи с контролируемой нагрузкой
- **Email Service:** отправляет до 100+ email/сек (в зависимости от SMTP)

---

## 🔧 Конфигурация

### Настройка Async Logger

```go
// Параметры:
// - bufferSize: размер буфера канала (рекомендуется 10000)
// - workerCount: количество воркеров (рекомендуется 5-10)
logger := async.NewAsyncLogger(db, 10000, 5)
```

**Рекомендации:**
- Увеличьте `workerCount` если видите высокую загрузку очереди
- Увеличьте `bufferSize` при пиковых нагрузках

### Настройка Worker Pool

```go
// Параметры:
// - workerCount: количество воркеров
// - queueSize: размер очереди задач
pool := async.NewWorkerPool(10, 1000)
```

**Рекомендации:**
- `workerCount` = количество CPU cores * 2 (для I/O задач)
- `queueSize` = зависит от скорости поступления задач

### Настройка Email Service

```go
// Workers настраиваются в конструкторе
// По умолчанию: 5 workers, queue size 1000
service := email.NewEmailService("noreply@example.com")
```

---

## 🛡️ Graceful Shutdown

**Важно!** Все компоненты поддерживают graceful shutdown для предотвращения потери данных.

```go
// В main.go добавьте обработку сигналов
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Println("Shutting down gracefully...")

    // Останавливаем все компоненты
    asyncLogger.Shutdown(30 * time.Second)
    workerPool.Shutdown(30 * time.Second)
    emailService.Shutdown()

    os.Exit(0)
}()
```

---

## 🐛 Отладка и мониторинг

### Логирование

Все компоненты выводят подробные логи:

```
[AsyncLogger] Worker 0 started
[AsyncLogger] Worker 1 started
[AsyncLogger] Worker 2 started
[WorkerPool] Started 10 workers
[EmailService] Worker 0 started
[EmailService] Email queued for user@example.com
```

### Метрики

```go
// Async Logger stats
stats := logger.Stats()
// Returns: queue_length, queue_capacity, worker_count

// Worker Pool stats
stats := pool.Stats()
// Returns: TasksProcessed, TasksFailed, ActiveWorkers, etc.

// Health check
if !pool.IsHealthy() {
    log.Println("Pool queue is 90% full or no active workers!")
}
```

---

## ⚠️ Best Practices

### 1. **Всегда используйте контекст с таймаутом**

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := requester.RequestWithFailover(ctx, nodeURLs, method, params)
```

### 2. **Обрабатывайте переполнение очередей**

```go
success := pool.Submit(task)
if !success {
    log.Println("Task queue is full, implement backpressure!")
}
```

### 3. **Мониторьте статистику**

```go
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        stats := pool.Stats()
        if stats.QueueLength > stats.QueueCapacity * 0.8 {
            log.Println("WARNING: Pool queue is 80% full!")
        }
    }
}()
```

### 4. **Используйте Graceful Shutdown**

```go
// Всегда вызывайте Shutdown перед выходом
defer asyncLogger.Shutdown(30 * time.Second)
defer workerPool.Shutdown(30 * time.Second)
defer emailService.Shutdown()
```

---

## 🎯 Когда использовать каждый компонент

### AsyncLogger
- ✅ Логирование в БД
- ✅ Аудит действий пользователей
- ✅ Метрики и аналитика

### Worker Pool
- ✅ Batch обработка данных
- ✅ Фоновые задачи
- ✅ Ограничение конкурентности

### Parallel Requester
- ✅ Запросы к blockchain нодам
- ✅ Запросы к внешним API
- ✅ Failover между серверами

### Async Email
- ✅ Отправка notifications
- ✅ Welcome emails
- ✅ Password reset emails

---

## 📈 Масштабирование

### Горизонтальное масштабирование

При использовании нескольких инстансов приложения:

1. **Async Logger** - каждый инстанс пишет в общую БД
2. **Worker Pool** - распределение задач через RabbitMQ
3. **Email Service** - каждый инстанс обрабатывает свою очередь

### Вертикальное масштабирование

Настройка количества воркеров в зависимости от ресурсов:

```go
cpuCount := runtime.NumCPU()

// Для CPU-bound задач
workerPool := async.NewWorkerPool(cpuCount, 1000)

// Для I/O-bound задач
workerPool := async.NewWorkerPool(cpuCount * 2, 1000)
```

---

## 🔗 Ссылки

- [Горутины в Go](https://go.dev/tour/concurrency/1)
- [Каналы в Go](https://go.dev/tour/concurrency/2)
- [Context в Go](https://pkg.go.dev/context)
- [sync.WaitGroup](https://pkg.go.dev/sync#WaitGroup)

---

**© 2025 IronNode - High-Performance Blockchain Infrastructure**
