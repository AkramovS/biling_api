# 🧪 Проверка проекта Biling API

Пошаговая инструкция для проверки работоспособности проекта.

---

## 📋 Предварительные требования

- **Go 1.21+** установлен
- **PostgreSQL 14+** установлен и запущен
- **psql** или **pgAdmin** для работы с БД

---

## 1️⃣ Создание базы данных

### Вариант A: Через командную строку

```bash
createdb biling_db
```

### Вариант B: Через psql

```bash
psql -U postgres

# В консоли psql:
CREATE DATABASE biling_db;
\q
```

---

## 2️⃣ Применение миграций

Выполните SQL файлы миграций по порядку:

```bash
# Создание таблиц
psql -U postgres -d biling_db -f migrations/000001_create_tables.up.sql

# Заполнение тестовыми данными
psql -U postgres -d biling_db -f migrations/000002_seed_data.up.sql
```

### Проверка создания таблиц

```bash
psql -U postgres -d biling_db

# В консоли psql:
\dt

# Должны увидеть таблицы:
# - system_accounts
# - system_group_info
# - system_rights
# - system_groups
# - users
# - accounts
# - users_accounts
# - tariffs
# - account_tariff_link
```

---

## 3️⃣ Создание тестового пользователя с правами

Добавим системного пользователя с полными правами доступа:

```sql
-- Подключение к БД
psql -U postgres -d biling_db

-- Создать тестового админа (пароль: password123)
-- Хеш для bcrypt "password123" (cost 12)
INSERT INTO system_accounts (login, password, name) 
VALUES ('admin', '$2a$12$VgF5pjJKJ9X5gKJ5Y3H0XOxLZvHXgKKZR5N9Q9JK5K5K5K5K5K5K5K', 'Администратор');

-- Создать группу "Администраторы"
INSERT INTO system_group_info (name, description) 
VALUES ('Администраторы', 'Полный доступ к системе');

-- Назначить права группе (все доступные FID)
INSERT INTO system_rights (group_id, fid) VALUES
    (1, 1),  -- FIDAccountsRead
    (1, 2),  -- FIDTariffsRead
    (1, 3);  -- FIDTariffsUpdate

-- Добавить пользователя в группу
INSERT INTO system_groups (group_id, user_id) 
VALUES (1, 1);

\q
```

### Генерация хеша пароля (опционально)

Если хотите создать свой хеш пароля:

```bash
cd tools
go run hash_password.go password123
```

---

## 4️⃣ Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```bash
# .env
PORT=4000
ENV=development
DB_DSN=postgres://postgres:postgres@localhost/biling_db?sslmode=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_MAX_IDLE_TIME=15m
JWT_SECRET=your-super-secret-jwt-key-change-in-production
```

**⚠️ ВАЖНО:** Замените `postgres:postgres` на ваши credentials PostgreSQL!

---

## 5️⃣ Установка зависимостей

```bash
cd d:\Work\Biling_api
go mod download
```

---

## 6️⃣ Запуск сервера

```bash
go run ./cmd/api
```

Вы должны увидеть:

```
database connection pool established
starting development server on :4000
```

---

## 7️⃣ Тестирование API

Откройте **новый терминал** и выполните следующие запросы:

### ✅ 1. Проверка healthcheck

```bash
curl http://localhost:4000/v1/health
```

**Ожидаемый ответ:**

```json
{
	"status": "available",
	"system_info": {
		"environment": "development",
		"version": "1.0.0"
	}
}
```

---

### ✅ 2. Вход в систему (получение JWT токена)

```bash
curl -X POST http://localhost:4000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"admin123"}'

```

**Ожидаемый ответ:**

```json
{
	"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
	"user": {
		"id": 1,
		"login": "admin",
		"created_at": "2024-01-01T00:00:00Z"
	}
}
```

**📝 Сохраните токен!** Он понадобится для следующих запросов.

---

### ✅ 3. Получение аккаунтов пользователя (требует FIDAccountsRead = 1)

**Windows :**

```bash
TOKEN="ваш_токен_сюда"
curl -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/users/1/accounts

```

**Ожидаемый ответ:**

```json
{
	"user": {
		"id": 1,
		"name": "Иван Петров"
	},
	"accounts": [
		{"id": 1},
		{"id": 2}
	]
}
```

---

### ✅ 4. Получение информации о тарифе аккаунта (требует FIDTariffsRead = 2)

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/account-tariffs/1

```

**Ожидаемый ответ:**

```json
{
	"account_tariff": {
		"id": 1,
		"account_id": 1,
		"tariff_id": 1,
		"version": 1,
		"updated_at": "2024-01-01T00:00:00Z",
		"updated_by": null,
		"updated_by_user": null
	}
}
```

---

### ✅ 5. Изменение тарифа (требует FIDTariffsUpdate = 3)

```bash
TOKEN="твой_токен_сюда"
curl -X PATCH http://localhost:4000/v1/account-tariffs/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"tariff_id\":2,\"version\":1}"
```

**Ожидаемый ответ (успех):**

```json
{
	"account_tariff": {
		"id": 1,
		"account_id": 1,
		"tariff_id": 2,
		"version": 2,
		"updated_at": "2024-01-01T12:34:56Z",
		"updated_by": 1,
		"updated_by_user": {
			"id": 1,
			"login": "admin"
		}
	}
}
```

---

### ✅ 6. Тест оптимистичной блокировки (конфликт версий)

Попробуйте изменить тариф с **устаревшей версией**:

```bash
TOKEN="твой_токен_сюда"
curl -X PATCH http://localhost:4000/v1/account-tariffs/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"tariff_id\":3,\"version\":1}"

```

**Ожидаемый ответ (409 Conflict):**

```json
{
	"error": {
		"code": "version_conflict",
		"message": "Record was modified by another user. Please review changes and retry.",
		"details": {
			"entity": "account_tariff_link",
			"id": 1
		}
	},
	"server": {
		"data": {
			"id": 1,
			"account_id": 1,
			"tariff_id": 2
		},
		"meta": {
			"version": 2,
			"updated_at": "2024-01-01T12:34:56Z",
			"updated_by": {...}
		}
	},
	"client": {
		"data": {
			"tariff_id": 3
		},
		"meta": {
			"expected_version": 1
		}
	}
}
```

---

### ❌ 7. Тест отсутствия прав доступа

Создайте пользователя **без прав**:

```sql
psql -U postgres -d biling_db

INSERT INTO system_accounts (login, password, name) 
VALUES ('username', '$2a$12$VgF5pjJKJ9X5gKJ5Y3H0XOxLZvHXgKKZR5N9Q9JK5K5K5K5K5K5K5K', 'Обычный пользователь');

-- НЕ добавляем его в группы!
\q
```

Войдите как этот пользователь:

```bash
curl -X POST http://localhost:4000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"login":"username","password":"password123"}'
```

Попробуйте получить аккаунты с новым токеном:

```bash
$TOKEN=новый_токен_user
curl -H "Authorization: Bearer $TOKEN" http://localhost:4000/v1/users/1/accounts
```

**Ожидаемый ответ (403 Forbidden):**

```json
{
	"error": "your user account doesn't have the necessary permissions to access this resource"
}
```

---

### Запуск с другим портом

```bash
go run ./cmd/api -port=8080
```
