# Инструкции по тестированию OrderFlow

## 🚀 Запуск системы

### 1. Запуск через Docker Compose

```bash
docker-compose up -d
```

### 2. Проверка статуса контейнеров

```bash
docker-compose ps
```

Все контейнеры должны быть в статусе "Up" и "healthy".

## 📋 API Endpoints

### Health Check

```bash
curl http://localhost:51193/health
```

Ожидаемый ответ: `{"status":"ok","timestamp":"..."}`

### Создание заказа

```bash
curl -X POST http://localhost:51193/api/orders \
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
```

Ожидаемый ответ:

```json
{
	"workflow_id": "order-processing-customer-001-...",
	"message": "Order processing started successfully"
}
```

### Проверка статуса заказа

```bash
curl "http://localhost:51193/api/orders/status?workflow_id=ORDER_WORKFLOW_ID"
```

### Проверка состояния workflow

```bash
curl "http://localhost:51193/api/orders/state?workflow_id=ORDER_WORKFLOW_ID"
```

### Отмена заказа

```bash
curl -X POST "http://localhost:51193/api/orders/cancel?workflow_id=ORDER_WORKFLOW_ID"
```

## 🖥️ Temporal Web UI

Откройте в браузере: http://localhost:8088

Здесь можно:

- Просматривать все workflow
- Отслеживать выполнение активностей
- Анализировать логи и ошибки
- Управлять workflow

## 📊 Мониторинг

### Логи приложения

```bash
docker-compose logs app -f
```

### Логи PostgreSQL

```bash
docker-compose logs postgres -f
```

### Логи Temporal

```bash
docker-compose logs temporal -f
```

## 🧪 Тестовые сценарии

### 1. Успешный заказ

1. Создайте заказ с доступным товаром
2. Проверьте, что workflow выполнился успешно
3. Убедитесь, что платеж обработан
4. Проверьте, что уведомление отправлено

### 2. Заказ с недоступным товаром

1. Создайте заказ с товаром, которого нет на складе
2. Проверьте, что workflow завершился с ошибкой
3. Убедитесь, что резервация отменена

### 3. Отмена заказа

1. Создайте заказ
2. Отмените его через API
3. Проверьте, что workflow остановлен
4. Убедитесь, что резервация отменена

## 🔧 Устранение неполадок

### Приложение не запускается

1. Проверьте логи: `docker-compose logs app`
2. Убедитесь, что PostgreSQL запущен: `docker-compose ps postgres`
3. Проверьте, что Temporal запущен: `docker-compose ps temporal`

### Ошибки в workflow

1. Откройте Temporal Web UI
2. Найдите workflow по ID
3. Просмотрите детали выполнения
4. Проверьте логи активностей

### Проблемы с базой данных

1. Проверьте подключение к PostgreSQL
2. Убедитесь, что миграции выполнены
3. Проверьте логи: `docker-compose logs postgres`

## 📝 Полезные команды

### Перезапуск системы

```bash
docker-compose down
docker-compose up -d --build
```

### Очистка данных

```bash
docker-compose down -v
docker-compose up -d
```

### Просмотр всех логов

```bash
docker-compose logs -f
```

### Проверка портов

```bash
netstat -an | grep 51193  # Приложение
netstat -an | grep 5432   # PostgreSQL
netstat -an | grep 7233   # Temporal
netstat -an | grep 8088   # Temporal Web UI
```
