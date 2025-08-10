# OrderFlow - Система обработки заказов с Temporal.io

Система обработки заказов, построенная с использованием Temporal.io для оркестрации бизнес-процессов. Реализована на Go с использованием PostgreSQL для хранения данных.

## 🏗️ Архитектура

Система построена по принципам **Domain-Driven Design (DDD)** и **Clean Architecture**:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP API Layer                           │
├─────────────────────────────────────────────────────────────┤
│                  Temporal Workflow                          │
├─────────────────────────────────────────────────────────────┤
│                    Activity Layer                           │
├─────────────────────────────────────────────────────────────┤
│                    Service Layer                            │
├─────────────────────────────────────────────────────────────┤
│                  Repository Layer                           │
├─────────────────────────────────────────────────────────────┤
│                   PostgreSQL                                │
└─────────────────────────────────────────────────────────────┘
```

### Основные компоненты:

- **Temporal Workflow** - оркестрация процесса обработки заказа
- **Activities** - отдельные шаги обработки (создание заказа, проверка склада, оплата, уведомления)
- **Services** - бизнес-логика для работы с заказами, складом, платежами, уведомлениями
- **Repositories** - слой доступа к данным PostgreSQL
- **HTTP API** - REST API для управления заказами

## 🚀 Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- Go 1.21+
- Make (опционально)

### 1. Клонирование и настройка

```bash
git clone <repository-url>
cd orderflow
```

### 2. Запуск инфраструктуры

```bash
# Запуск PostgreSQL и Temporal
make docker-up

# Или вручную:
docker-compose up -d
```

### 3. Запуск приложения

```bash
# Запуск в режиме разработки (с автоматическим запуском инфраструктуры)
make dev

# Или вручную:
go run cmd/main.go
```

### 4. Проверка работоспособности

```bash
# Проверка здоровья приложения
make health

# Создание тестового заказа
make create-order

# Открытие Temporal Web UI
make temporal-ui
```

## 📋 API Endpoints

### Создание заказа

```bash
POST /api/orders
Content-Type: application/json

{
  "customer_id": "customer-001",
  "items": [
    {
      "product_id": "prod-001",
      "name": "iPhone 15 Pro",
      "quantity": 1,
      "price": 999.99
    }
  ]
}
```

### Получение статуса заказа

```bash
GET /api/orders/status?workflow_id=<workflow_id>
```

### Отмена заказа

```bash
POST /api/orders/cancel?workflow_id=<workflow_id>
```

### Получение состояния workflow

```bash
GET /api/orders/state?workflow_id=<workflow_id>
```

### Проверка здоровья

```bash
GET /health
```

## 🔄 Процесс обработки заказа

1. **Создание заказа** - создание записи в БД
2. **Проверка склада** - проверка наличия товаров и резервирование
3. **Обработка платежа** - симуляция платежной системы
4. **Подтверждение заказа** - подтверждение резервирования товаров
5. **Уведомление клиента** - отправка уведомления об успешном заказе

### Обработка ошибок

- **Недостаточно товаров** - заказ отменяется, резервирование освобождается
- **Ошибка платежа** - заказ отменяется, резервирование освобождается
- **Ошибка уведомления** - заказ остается активным, но клиент не уведомлен

## 🛠️ Разработка

### Структура проекта

```
orderflow/
├── cmd/
│   └── main.go                 # Точка входа приложения
├── internal/
│   ├── adapter/
│   │   └── repository/         # PostgreSQL репозитории
│   ├── domain/                 # Доменные модели и интерфейсы
│   │   ├── inventory/          # Склад
│   │   ├── notification/       # Уведомления
│   │   ├── order/              # Заказы
│   │   ├── payment/            # Платежи
│   │   └── workflow/           # Temporal workflow
│   ├── handlers/               # HTTP handlers
│   ├── httpserver/             # HTTP сервер
│   └── usecase/
│       ├── activity/           # Temporal activities
│       └── service/            # Бизнес-сервисы
├── migrations/                 # SQL миграции
├── pkg/                        # Общие пакеты
└── docker-compose.yaml         # Инфраструктура
```

### Команды для разработки

```bash
# Установка зависимостей
make deps

# Форматирование кода
make fmt

# Проверка линтером
make lint

# Запуск тестов
make test

# Полная проверка кода
make check

# Сборка приложения
make build
```

### Добавление новых функций

1. **Новый домен**: создайте папку в `internal/domain/`
2. **Новый сервис**: создайте файл в `internal/usecase/service/`
3. **Новый репозиторий**: создайте файл в `internal/adapter/repository/`
4. **Новая Activity**: создайте файл в `internal/usecase/activity/`
5. **Новый HTTP handler**: создайте файл в `internal/handlers/`

## 📊 Мониторинг

### Temporal Web UI

- URL: http://localhost:8088
- Просмотр workflow, activities, history

### Логи

```bash
# Логи приложения
make logs

# Логи инфраструктуры
make docker-logs
```

## 🧪 Тестирование

### Unit тесты

```bash
make test
```

### Интеграционные тесты

```bash
# Запуск с реальной БД
docker-compose up -d postgres
go test -tags=integration ./...
```

### Примеры использования

```bash
# Создание заказа
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-001",
    "items": [
      {
        "product_id": "prod-001",
        "name": "iPhone 15 Pro",
        "quantity": 1,
        "price": 999.99
      }
    ]
  }'

# Получение статуса (замените WORKFLOW_ID на реальный ID)
curl -X GET "http://localhost:8080/api/orders/status?workflow_id=order-processing-customer-001-1234567890"

# Отмена заказа
curl -X POST "http://localhost:8080/api/orders/cancel?workflow_id=order-processing-customer-001-1234567890"
```

## 🔧 Конфигурация

### Переменные окружения

```bash
# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=orderflow
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password

# Temporal
TEMPORAL_HOST=localhost
TEMPORAL_PORT=7233

# HTTP Server
HTTP_PORT=8080
```

### Настройка БД

Миграции применяются автоматически при запуске PostgreSQL. Демо-данные (товары) также загружаются автоматически.

## 🚨 Troubleshooting

### Проблемы с подключением к PostgreSQL

```bash
# Проверка статуса контейнера
docker-compose ps postgres

# Проверка логов
docker-compose logs postgres
```

### Проблемы с Temporal

```bash
# Проверка статуса контейнера
docker-compose ps temporal

# Проверка логов
docker-compose logs temporal
```

### Проблемы с приложением

```bash
# Проверка логов приложения
# Логи выводятся в stdout при запуске через go run
```

## 📝 Лицензия

MIT License

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request
