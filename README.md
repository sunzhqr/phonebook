# Phonebook API

**Phonebook API** — сервис управления контактами (CRUD + поиск) с поддержкой хранения нескольких телефонов на контакт, нормализацией номеров и полнотекстовым поиском.  
Проект построен на Go с чистой архитектурой: `handler → service → repository`.

---

## Архитектура

```text
cmd/server           — точка входа (HTTP-сервер)
internal/
  handler/           — HTTP-эндпоинты (REST, валидация входных данных)
  service/           — бизнес-логика (нормализация телефонов, правила первичности)
  repository/        — слой доступа к данным (PostgreSQL, pgx)
  httpserver/        — конфигурируемый chi.Server с middleware
  logger/            — обёртка над zap
  config/            — загрузка конфигурации (.env)
pkg/
  normalizer/        — утилита нормализации телефонов (E.164, digits-only)
```

**Слои:**
- **Handler** — принимает HTTP-запрос, валидирует, вызывает сервис, формирует ответ.
- **Service** — применяет бизнес-правила, преобразует DTO в структуры репозитория, нормализует данные.
- **Repository** — работает с БД, возвращает модели, управляет транзакциями.
- **Adapter** (в перспективе) — интеграция с внешними сервисами (например, валидация номеров, SMS API).
- **Pkg** — переиспользуемые пакеты (независимы от бизнес-логики).

---

## Технологии

- **Go 1.24**
- **PostgreSQL 16** (расширение `pg_trgm` для поиска)
- **pgx v5** — высокопроизводительный драйвер PostgreSQL
- **chi v5** — роутер с middleware
- **zap** — структурированное логирование
---

## Запуск

### 1. Клонировать репозиторий
```bash
git clone https://github.com/sunzhqr/phonebook.git
cd phonebook
```

### 2. Настроить окружение
Создайте `.env` по образцу `.env-example`

### 3. Поднять PostgreSQL (Docker)
```bash
docker-compose up
```
или
```bash
docker run --name phonebook-pg   -e POSTGRES_PASSWORD=postgres   -e POSTGRES_DB=phonebook   -p 5432:5432   -d postgres:16
```

### 4. Применить миграции
```bash
make migrate-up
```

### 5. Запустить сервер
```bash
make run
# или
go run ./cmd/server
```

---

## API

### Создать контакт
```http
POST /api/v1/contacts
Content-Type: application/json

{
  "first_name": "Sanzhar",
  "last_name": "Sanzharov",
  "company": "Forte",
  "phones": [
    {"label": "work", "phone_raw": "+77711234567", "is_primary": true},
    {"label": "home", "phone_raw": "+77021234567"}
  ]
}
```

### Обновить контакт
```http
PUT /api/v1/contacts/{id}
```

### Получить контакт
```http
GET /api/v1/contacts/{id}
```

### Удалить контакт
```http
DELETE /api/v1/contacts/{id}
```

### Поиск
```http
GET /api/v1/contacts/search?q=+7771
```

---

## Тестирование

### Unit-тесты
```bash
make test
```
или
```bash
go test ./... -v
```
