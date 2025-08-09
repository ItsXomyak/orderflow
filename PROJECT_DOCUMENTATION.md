# 📚 OrderFlow - Документация проекта

## 🎯 Что это за проект?

**OrderFlow** - это система обработки заказов, построенная на Go с использованием Temporal.io для оркестрации рабочих процессов. Проект демонстрирует современную архитектуру микросервисов с использованием Docker и PostgreSQL.

## 🏗️ Архитектура системы

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Interface │    │   API Server    │    │  Temporal      │
│   (Port 8080)   │◄──►│   (Port 8080)   │◄──►│  (Port 7233)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   PostgreSQL    │    │     Redis       │
                       │   (Port 5432)   │    │   (Port 6379)   │
                       └─────────────────┘    └─────────────────┘
```

### Компоненты системы:

- **API Server** - REST API для управления заказами
- **Worker** - обработчик рабочих процессов Temporal
- **Temporal** - оркестратор рабочих процессов
- **PostgreSQL** - основная база данных
- **Redis** - кэш для Temporal
- **Web UI** - веб-интерфейс для управления

## 📁 Структура проекта

```
orderflow/
├── 📁 cmd/                    # Точки входа приложения
│   ├── 📁 api/               # API сервер
│   └── 📁 worker/            # Worker для Temporal
├── 📁 internal/               # Внутренняя логика приложения
│   ├── 📁 domain/            # Бизнес-логика и модели
│   ├── 📁 application/       # Слой приложения (use cases)
│   ├── 📁 interfaces/        # HTTP handlers, репозитории
│   └── 📁 infrastructure/    # Внешние зависимости (БД, кэш)
├── 📁 migrations/             # SQL миграции для PostgreSQL
├── 📁 temporal-config/        # Конфигурация Temporal
├── 📁 templates/              # HTML шаблоны
├── 📁 static/                 # Статические файлы (CSS, JS)
├── 📁 bin/                    # Скомпилированные бинарники
├── 📁 pkg/                    # Переиспользуемые пакеты
├── 🐳 Dockerfile              # Docker образ
├── 🐳 docker-compose.yml      # Оркестрация сервисов
├── 📝 env.example             # Пример переменных окружения
├── 📝 go.mod                  # Go модули и зависимости
├── 📝 Makefile                # Команды для разработки
└── 📝 README.md               # Основная документация
```

## 🔧 Что мы добавили/изменили

### 1. Файл переменных окружения (`env.example`)

**Зачем нужен:** Вынесли все хардкод значения из Docker в переменные окружения для гибкости настройки.

**Что содержит:**

```bash
# PostgreSQL настройки
POSTGRES_DB=orderflow
POSTGRES_USER=orderflow
POSTGRES_PASSWORD=orderflow123
POSTGRES_PORT=5432
POSTGRES_HOST=postgres

# Redis настройки
REDIS_PORT=6379

# Temporal настройки
TEMPORAL_HOST_PORT=temporal:7233
TEMPORAL_WEB_PORT=8088

# Порт приложения
API_PORT=8080
```

### 2. Обновленный `docker-compose.yml`

**Что изменилось:** Все значения теперь используют переменные окружения `${VARIABLE_NAME}`

**Пример изменений:**

```yaml
# Было:
POSTGRES_PASSWORD: orderflow123

# Стало:
POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
```

**Преимущества:**

- ✅ Легко менять настройки без редактирования Docker файлов
- ✅ Разные конфигурации для разных окружений (dev, staging, prod)
- ✅ Безопасность - пароли не в коде
- ✅ Гибкость настройки портов

## 🚀 Как запустить проект

### Быстрый старт:

1. **Скопируйте переменные окружения:**

   ```bash
   cp env.example .env
   ```

2. **Запустите все сервисы:**

   ```bash
   docker-compose up -d
   ```

3. **Проверьте статус:**
   ```bash
   docker-compose ps
   ```

### Доступные сервисы:

| Сервис           | URL                   | Описание              |
| ---------------- | --------------------- | --------------------- |
| **API**          | http://localhost:8080 | Основное приложение   |
| **Temporal Web** | http://localhost:8088 | Управление процессами |
| **PostgreSQL**   | localhost:5432        | База данных           |
| **Redis**        | localhost:6379        | Кэш                   |

## 🛠️ Разработка

### Локальная разработка:

```bash
# Установка зависимостей
go mod download

# Запуск API
go run cmd/api/main.go

# Запуск Worker
go run cmd/worker/main.go
```

### Docker разработка:

```bash
# Пересборка контейнеров
docker-compose build

# Просмотр логов
docker-compose logs -f api
docker-compose logs -f worker
```

### Полезные команды Make:

```bash
make start-full-stack    # Запуск всех сервисов
make docker-logs         # Просмотр логов
make docker-down         # Остановка сервисов
make build               # Сборка Go приложения
```

## 🔐 Безопасность

### Важные моменты:

1. **Файл `.env` НЕ коммитится** в Git (уже добавлен в `.gitignore`)
2. **Пароли по умолчанию** только для разработки
3. **В продакшене** обязательно измените все пароли

### Настройка продакшена:

```bash
# Создайте .env.production
cp env.example .env.production

# Отредактируйте с реальными значениями
nano .env.production

# Запустите с продакшен конфигом
docker-compose --env-file .env.production up -d
```

## 📊 Мониторинг и логи

### Просмотр логов:

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f api
docker-compose logs -f postgres
docker-compose logs -f temporal
```

### Health checks:

Все сервисы имеют встроенные проверки здоровья:

- PostgreSQL: проверка готовности БД
- Redis: ping команда
- API/Worker: автоматический перезапуск при сбоях

## 🔄 Workflow с Temporal

### Как это работает:

1. **API получает заказ** от пользователя
2. **Создается Temporal workflow** для обработки
3. **Worker выполняет шаги** workflow (валидация, оплата, доставка)
4. **Каждый шаг** может быть повторен при ошибке
5. **Результат сохраняется** в PostgreSQL

### Преимущества Temporal:

- ✅ **Надежность** - автоматические повторы при сбоях
- ✅ **Мониторинг** - веб-интерфейс для отслеживания процессов
- ✅ **Масштабируемость** - легко добавлять новые workers
- ✅ **Версионирование** - безопасное обновление workflow

## 🐛 Отладка

### Частые проблемы:

1. **Порт занят:**

   ```bash
   # Проверьте что использует порт
   netstat -tulpn | grep :8080
   ```

2. **Контейнер не запускается:**

   ```bash
   # Посмотрите логи
   docker-compose logs service_name
   ```

3. **БД не подключается:**
   ```bash
   # Проверьте переменные в .env
   cat .env | grep POSTGRES
   ```

### Полезные команды:

```bash
# Перезапуск сервиса
docker-compose restart api

# Полная пересборка
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## 📈 Следующие шаги

### Что можно улучшить:

1. **Добавить мониторинг** (Prometheus + Grafana)
2. **Настроить CI/CD** pipeline
3. **Добавить тесты** для API и workflow
4. **Реализовать аутентификацию** пользователей
5. **Добавить метрики** производительности

### Полезные ссылки:

- [Temporal Documentation](https://docs.temporal.io/)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Docker Compose Reference](https://docs.docker.com/compose/)

---

## 💡 Заключение

Теперь у вас есть полностью настроенная система OrderFlow с:

- ✅ Вынесенными в переменные окружения настройками
- ✅ Готовыми Docker контейнерами
- ✅ Temporal для надежной обработки заказов
- ✅ PostgreSQL для хранения данных
- ✅ Простым запуском через `docker-compose up -d`

Просто скопируйте `env.example` в `.env`, отредактируйте значения под ваше окружение и запускайте! 🚀
