# Biling API

JWT-авторизация с RBAC (Role-Based Access Control) следуя паттернам Alex Edwards.

## Требования

- Go 1.21+
- PostgreSQL 14+

## Быстрый старт

### 1. Создать базу данных

```bash
createdb utm
```

Или через psql:
```sql
CREATE DATABASE utm;
```

### 2. Применить миграции

Вручную выполните SQL из файлов миграций:

```bash
psql utm < migrations/000001_create_tables.up.sql
psql utm < migrations/000002_seed_data.up.sql
```

### 3. Запустить сервер

```bash
go run ./cmd/api
```

Сервер запустится на `http://localhost:4000`

## API Endpoints

### Публичные эндпоинты

- `GET /v1/health` - Проверка состояния
- `POST /v1/auth/register` - Регистрация пользователя
- `POST /v1/auth/login` - Вход (получение JWT токена)

### Защищенные эндпоинты

- `GET /v1/users/:id/accounts` - Получить лицевые счета пользователя (требуется права `accounts:read`)

## Примеры использования

### Регистрация

```bash
curl -X POST http://localhost:4000/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"newuser@example.com","password":"password123"}'
```

### Вход (admin - с правами доступа)

```bash
curl -X POST http://localhost:4000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}'
```

Ответ:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "admin@example.com",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### Вход (user - без прав доступа)

```bash
curl -X POST http://localhost:4000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Получение лицевых счетов (с токеном admin)

```bash
curl http://localhost:4000/v1/users/1/accounts \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

Успешный ответ (для admin):
```json
{
  "user": {
    "id": 1,
    "name": "Иван Иванов"
  },
  "accounts": [
    {
      "id": 1,
      "name": "Лицевой счет 10001"
    },
    {
      "id": 2,
      "name": "Лицевой счет 10002"
    },
    {
      "id": 3,
      "name": "Лицевой счет 10003"
    }
  ]
}
```

Ошибка (для user без прав):
```json
{
  "error": "your user account doesn't have the necessary permissions to access this resource"
}
```

## Тестовые данные

### Системные пользователи (auth_users)

1. **admin@example.com** / password123
   - Состоит в группе `api_users`
   - Имеет доступ к `/v1/users/:id/accounts`

2. **user@example.com** / password123
   - НЕ состоит в группе
   - НЕ имеет доступа к защищенным эндпоинтам

### Бизнес-пользователи (users)

- 5 пользователей с именами
- 10 лицевых счетов
- Связи через `users_accounts`

## Структура проекта

```
.
├── cmd/
│   └── api/              # Главное приложение
│       ├── main.go       # Точка входа
│       ├── routes.go     # Роуты
│       ├── middleware.go # Аутентификация/авторизация
│       ├── helpers.go    # Вспомогательные функции
│       ├── errors.go     # Обработка ошибок
│       └── *_handlers.go # Обработчики запросов
├── internal/
│   ├── data/            # Модели данных
│   │   ├── models.go
│   │   ├── users.go
│   │   ├── accounts.go
│   │   ├── auth_users.go
│   │   ├── groups.go
│   │   └── tokens.go
│   └── validator/       # Валидация
│       └── validator.go
├── migrations/          # SQL миграции
└── go.mod
```

## Конфигурация

Переменные окружения и флаги:

```bash
go run ./cmd/api \
  -port=4000 \
  -env=development \
  -db-dsn="postgres://postgres:postgres@localhost/utm?sslmode=disable" \
  -jwt-secret="your-secret-key-change-this-in-production"
```

## RBAC (Role-Based Access Control)

Система использует группы и разрешения:

- **Группы** (groups): Коллекции пользователей
- **Разрешения** (group_permissions): Права доступа (resource:action)
- **Членство** (group_members): Связь пользователей с группами

Например:
- Группа: `api_users`
- Разрешение: `accounts:read`
- Член: `admin@example.com`

## Безопасность

- Пароли хешируются с помощью bcrypt (cost 12)
- JWT токены подписываются HMAC-SHA256
- Токены действительны 24 часа
- Middleware проверяет токен и права доступа для каждого запроса
